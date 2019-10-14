package dgopcore

import (
	"errors"
	"fmt"

	"github.com/lyraproj/dgo/dgo"
	"github.com/lyraproj/dgo/typ"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
)

// ToPcore converts a dgo.Value into its corresponding px.Value.
func ToPcore(v dgo.Value) px.Value {
	var cv px.Value
	switch v := v.(type) {
	case dgo.String:
		cv = types.WrapString(v.GoString())
	case dgo.Integer:
		cv = types.WrapInteger(v.GoInt())
	case dgo.Float:
		cv = types.WrapFloat(v.GoFloat())
	case dgo.Boolean:
		cv = types.WrapBoolean(v.GoBool())
	case dgo.Array:
		cv = toArray(v)
	case dgo.Map:
		cv = toMap(v)
	case dgo.Binary:
		cv = types.WrapBinary(v.GoBytes())
	case dgo.Native:
		cv = toNative(v)
	case dgo.Regexp:
		cv = types.WrapRegexp2(v.GoRegexp())
	case dgo.Sensitive:
		cv = types.WrapSensitive(ToPcore(v.Unwrap()))
	case dgo.Time:
		cv = types.WrapTimestamp(v.GoTime())
	case dgo.Nil:
		cv = px.Undef
	case dgo.Type:
		cv = toType(v)
	default:
		panic(fmt.Errorf(`unable to create a px.Value from a dgo %s`, v.Type()))
	}
	return cv
}

func toArray(a dgo.Array) px.List {
	ps := make([]px.Value, a.Len())
	a.EachWithIndex(func(v dgo.Value, i int) { ps[i] = ToPcore(v) })
	return types.WrapValues(ps)
}

func toNative(v dgo.Native) *types.RuntimeValue {
	return types.WrapRuntime(v.GoValue())
}

func toMap(v dgo.Map) px.OrderedMap {
	hs := make([]*types.HashEntry, 0, v.Len())
	v.EachEntry(func(e dgo.MapEntry) { hs = append(hs, types.WrapHashEntry(ToPcore(e.Key()), ToPcore(e.Value()))) })
	return types.WrapHash(hs)
}

func toTypes(ts dgo.Array) []px.Type {
	ps := make([]px.Type, ts.Len())
	ts.EachWithIndex(func(v dgo.Value, i int) { ps[i] = toType(v.(dgo.Type)) })
	return ps
}

func toIntegerType(min, max int) *types.IntegerType {
	return types.NewIntegerType(int64(min), int64(max))
}

func toType(t dgo.Type) (pt px.Type) {
	if f, ok := toTypeFuncs[t.TypeIdentifier()]; ok {
		return f(t)
	}
	panic(fmt.Errorf(`unable to create pcore Type from a '%s'`, t))
}

func toStruct(st dgo.StructMapType) px.Type {
	if st.Additional() {
		panic(errors.New(`unable to create pcore Struct that allows additional entries`))
	}
	es := make([]*types.StructElement, st.Len())
	i := 0
	st.Each(func(e dgo.StructMapEntry) {
		kt := toType(e.Key().(dgo.Type))
		if !e.Required() {
			kt = types.NewOptionalType(kt)
		}
		es[i] = types.NewStructElement(kt, toType(e.Value().(dgo.Type)))
		i++
	})
	return types.NewStructType(es)
}

type toTypeFunc func(dgo.Type) px.Type

var toTypeFuncs map[dgo.TypeIdentifier]toTypeFunc

func init() {
	toTypeFuncs = map[dgo.TypeIdentifier]toTypeFunc{
		dgo.TiBoolean: func(t dgo.Type) px.Type {
			return types.DefaultBooleanType()
		},
		dgo.TiTrue: func(t dgo.Type) px.Type {
			return types.NewBooleanType(true)
		},
		dgo.TiFalse: func(t dgo.Type) px.Type {
			return types.NewBooleanType(false)
		},
		dgo.TiFloat: func(t dgo.Type) px.Type {
			return types.DefaultFloatType()
		},
		dgo.TiFloatExact: func(t dgo.Type) px.Type {
			rt := t.(dgo.FloatRangeType)
			return types.NewFloatType(rt.Min(), rt.Max())
		},
		dgo.TiFloatRange: func(t dgo.Type) px.Type {
			rt := t.(dgo.FloatRangeType)
			if !rt.Inclusive() {
				panic(errors.New(`unable to create non inclusive pcore Float type`))
			}
			return types.NewFloatType(rt.Min(), rt.Max())
		},
		dgo.TiInteger: func(t dgo.Type) px.Type {
			return types.DefaultIntegerType()
		},
		dgo.TiIntegerExact: func(t dgo.Type) px.Type {
			rt := t.(dgo.IntegerRangeType)
			return types.NewIntegerType(rt.Min(), rt.Max())
		},
		dgo.TiIntegerRange: func(t dgo.Type) px.Type {
			rt := t.(dgo.IntegerRangeType)
			mx := rt.Max()
			if !rt.Inclusive() {
				mx--
			}
			return types.NewIntegerType(rt.Min(), mx)
		},
		dgo.TiString: func(t dgo.Type) px.Type {
			return types.DefaultStringType()
		},
		dgo.TiStringExact: func(t dgo.Type) px.Type {
			return types.NewStringType(nil, t.(dgo.ExactType).Value().(dgo.String).GoString())
		},
		dgo.TiStringSized: func(t dgo.Type) px.Type {
			st := t.(dgo.StringType)
			return types.NewStringType(toIntegerType(st.Min(), st.Max()), ``)
		},
		dgo.TiStringPattern: func(t dgo.Type) px.Type {
			rx := t.(dgo.ExactType).Value().(dgo.Regexp)
			return types.NewPatternType([]*types.RegexpType{types.NewRegexpTypeR(rx.GoRegexp())})
		},
		dgo.TiCiString: func(t dgo.Type) px.Type {
			v := t.(dgo.ExactType).Value().(dgo.String)
			return types.NewEnumType([]string{v.GoString()}, true)
		},
		dgo.TiAny: func(t dgo.Type) px.Type {
			return types.DefaultAnyType()
		},
		dgo.TiAnyOf: func(t dgo.Type) px.Type {
			return types.NewVariantType(toTypes(t.(dgo.TernaryType).Operands())...)
		},
		dgo.TiArray: func(t dgo.Type) px.Type {
			at := t.(dgo.ArrayType)
			return types.NewArrayType(toType(at.ElementType()), toIntegerType(at.Min(), at.Max()))
		},
		dgo.TiArrayExact: func(t dgo.Type) px.Type {
			return toArray(t.(dgo.ExactType).Value().(dgo.Array)).PType()
		},
		dgo.TiTuple: func(t dgo.Type) px.Type {
			es := toTypes(t.(dgo.TupleType).ElementTypes())
			tc := len(es)
			if tc == 0 {
				return types.DefaultTupleType()
			}
			return types.NewTupleType(es, toIntegerType(tc, tc))
		},
		dgo.TiMap: func(t dgo.Type) px.Type {
			mt := t.(dgo.MapType)
			return types.NewHashType(toType(mt.KeyType()), toType(mt.ValueType()), toIntegerType(mt.Min(), mt.Max()))
		},
		dgo.TiMapExact: func(t dgo.Type) px.Type {
			return toMap(t.(dgo.ExactType).Value().(dgo.Map)).PType()
		},
		dgo.TiStruct: func(t dgo.Type) px.Type {
			return toStruct(t.(dgo.StructMapType))
		},
		dgo.TiBinary: func(t dgo.Type) px.Type {
			return types.DefaultBinaryType()
		},
		dgo.TiMeta: func(t dgo.Type) px.Type {
			ct := t.(dgo.UnaryType).Operand()
			var pt px.Type
			if ct == nil {
				pt = types.DefaultTypeType().PType()
			} else {
				pt = toType(ct)
			}
			return types.NewTypeType(pt)
		},
		dgo.TiNative: func(t dgo.Type) px.Type {
			rt := t.(dgo.NativeType).GoType()
			if rt == nil {
				return types.DefaultRuntimeType()
			}
			return types.NewGoRuntimeType(rt)
		},
		dgo.TiRegexp: func(t dgo.Type) px.Type {
			return types.DefaultRegexpType()
		},
		dgo.TiRegexpExact: func(t dgo.Type) px.Type {
			return types.NewRegexpTypeR(t.(dgo.ExactType).Value().(dgo.Regexp).GoRegexp())
		},
		dgo.TiSensitive: func(t dgo.Type) px.Type {
			if op := t.(dgo.UnaryType).Operand(); typ.Any != op {
				return types.NewSensitiveType(toType(op))
			}
			return types.DefaultSensitiveType()
		},
		dgo.TiTime: func(t dgo.Type) px.Type {
			return types.DefaultTimestampType()
		},
		dgo.TiTimeExact: func(t dgo.Type) px.Type {
			tm := t.(dgo.ExactType).Value().(dgo.Time).GoTime()
			return types.NewTimestampType(tm, tm)
		},
	}
}
