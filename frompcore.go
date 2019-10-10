// Package dgopcore contains the two function necessary to convert a dgo.Value into a pcore.Value and
// vice versa.
package dgopcore

import (
	"fmt"
	"time"

	"github.com/lyraproj/dgo/dgo"
	"github.com/lyraproj/dgo/newtype"
	"github.com/lyraproj/dgo/typ"
	"github.com/lyraproj/dgo/vf"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
)

// FromPcore converts a pcore.Value into its corresponding dgo.Value
func FromPcore(pv px.Value) dgo.Value {
	if pv.Equals(px.Undef, nil) {
		return vf.Nil
	}
	var v dgo.Value
	switch pv := pv.(type) {
	case px.StringValue:
		v = vf.String(pv.String())
	case px.Integer:
		v = vf.Integer(pv.Int())
	case px.Float:
		v = vf.Float(pv.Float())
	case px.Boolean:
		v = vf.Boolean(pv.Bool())
	case px.OrderedMap:
		v = fromMap(pv)
	case *types.Binary:
		v = vf.Binary(pv.Bytes(), false)
	case px.List:
		v = fromList(pv)
	case *types.Regexp:
		v = vf.Regexp(pv.Regexp())
	case *types.Sensitive:
		v = vf.Sensitive(FromPcore(pv.Unwrap()))
	case *types.Timestamp:
		v = vf.Time(*(*time.Time)(pv))
	case px.Type:
		v = fromType(pv)
	default:
		panic(fmt.Errorf(`unable to create a dgo.Value from a pcore %s`, pv.PType().Name()))
	}
	return v
}

func fromList(pv px.List) dgo.Array {
	vs := make([]dgo.Value, pv.Len())
	pv.EachWithIndex(func(v px.Value, i int) {
		vs[i] = FromPcore(v)
	})
	return vf.Array(vs)
}

func fromMap(pv px.OrderedMap) dgo.Map {
	m := vf.MutableMap(nil)
	pv.Each(func(v px.Value) {
		e := v.(px.MapEntry)
		m.Put(FromPcore(e.Key()), FromPcore(e.Value()))
	})
	return m
}

func fromType(pt px.Type) dgo.Type {
	if vt, ok := fromPrimitiveType(pt); ok {
		return vt
	}
	if vt, ok := fromCollectionType(pt); ok {
		return vt
	}
	if vt, ok := fromComplexType(pt); ok {
		return vt
	}
	panic(fmt.Errorf(`unable to create a dgo.Type from a pcore %s`, pt.Name()))
}

func fromComplexType(pt px.Type) (vt dgo.Type, ok bool) {
	ok = true
	switch pt := pt.(type) {
	case *types.AnyType:
		vt = typ.Any
	case *types.BinaryType:
		vt = typ.Binary
	case *types.EnumType:
		vt = fromEnumType(pt)
	case *types.TypeType:
		vt = fromType(pt.ContainedType()).Type()
	case *types.NotUndefType:
		vt = newtype.AllOf(newtype.Not(typ.Nil), fromType(pt.ContainedType()))
	case *types.OptionalType:
		vt = fromType(pt.ContainedType())
		if !vt.Assignable(typ.Nil) {
			vt = newtype.AnyOf(typ.Nil, vt)
		}
	case *types.PatternType:
		vt = fromPatternType(pt)
	case *types.RegexpType:
		vt = fromRegexpType(pt)
	case *types.SensitiveType:
		vt = fromSensitiveType(pt)
	case *types.TimestampType:
		vt = fromTimestampType(pt)
	case *types.VariantType:
		vt = fromVariantType(pt.Types()...)
	default:
		ok = false
	}
	return
}

func fromPrimitiveType(pt px.Type) (vt dgo.Type, ok bool) {
	ok = true
	switch pt := pt.(type) {
	case *types.BooleanType:
		vt = fromBooleanType(pt)
	case *types.NumericType:
		vt = newtype.AnyOf(typ.Float, typ.Integer)
	case *types.FloatType:
		vt = newtype.FloatRange(pt.Min(), pt.Max(), true)
	case *types.IntegerType:
		vt = newtype.IntegerRange(pt.Min(), pt.Max(), true)
	case px.StringType:
		sz := pt.Size().(*types.IntegerType)
		vt = newtype.String(sz.Min(), sz.Max())
	case *types.UndefType:
		vt = typ.Nil
	default:
		ok = false
	}
	return
}

func fromCollectionType(pt px.Type) (vt dgo.Type, ok bool) {
	ok = true
	switch pt := pt.(type) {
	case *types.ArrayType:
		vt = fromArrayType(pt)
	case *types.TupleType:
		vt = fromTupleType(pt)
	case *types.CollectionType:
		vt = newtype.AnyOf(typ.Array, typ.Map)
	case *types.HashType:
		vt = fromHashType(pt)
	case *types.StructType:
		vt = fromStructType(pt)
	default:
		ok = false
	}
	return
}

func fromArrayType(pt *types.ArrayType) dgo.Type {
	if types.DefaultArrayType().Equals(pt, nil) {
		return typ.Array
	}
	sz := pt.Size()
	return newtype.Array(fromType(pt.ElementType()), sz.Min(), sz.Max())
}

func fromBooleanType(pt *types.BooleanType) dgo.Type {
	vt := typ.Boolean
	if pt.IsInstance(types.BooleanTrue, nil) {
		if !pt.IsInstance(types.BooleanFalse, nil) {
			vt = typ.True
		}
	} else {
		vt = typ.False
	}
	return vt
}

func fromHashType(pt *types.HashType) dgo.Type {
	if types.DefaultHashType().Equals(pt, nil) {
		return typ.Map
	}
	sz := pt.Size()
	return newtype.Map(fromType(pt.KeyType()), fromType(pt.ValueType()), sz.Min(), sz.Max())
}

func fromEnumType(pt *types.EnumType) dgo.Type {
	if types.DefaultEnumType().Equals(pt, nil) {
		// A default Enum in pcore accepts all Enum's, so it's actually an unconstrained String
		return typ.String
	}
	es := pt.Strings()
	if pt.IsCaseInsensitive() {
		return newtype.CiEnum(es...)
	}
	return newtype.Enum(es...)
}

func fromPatternType(pt *types.PatternType) dgo.Type {
	if types.DefaultPatternType().Equals(pt, nil) {
		return typ.String
	}
	patterns := pt.Patterns()
	tps := make([]dgo.Type, patterns.Len())
	patterns.EachWithIndex(func(v px.Value, i int) {
		tps[i] = newtype.Pattern(patterns.At(0).(*types.RegexpType).Regexp())
	})
	return newtype.AnyOf(tps...)
}

func fromRegexpType(pt *types.RegexpType) dgo.Type {
	if types.DefaultRegexpType().Equals(pt, nil) {
		return typ.Regexp
	}
	return vf.Regexp(pt.Regexp()).Type()
}

func fromSensitiveType(pt *types.SensitiveType) dgo.Type {
	if types.DefaultSensitiveType().Equals(pt, nil) {
		return typ.Sensitive
	}
	return newtype.Sensitive(fromType(pt.ContainedType()))
}

func fromStructType(pt *types.StructType) dgo.Type {
	if types.DefaultStructType().Equals(pt, nil) {
		return typ.Map
	}
	pes := pt.Elements()
	ds := make([]dgo.StructEntry, len(pes))
	for i := range pes {
		pe := pes[i]
		ds[i] = newtype.StructEntry(pe.Name(), fromType(pe.Value()), !pe.Optional())
	}
	return newtype.Struct(false, ds...)
}

func fromTimestampType(pt *types.TimestampType) dgo.Type {
	if types.DefaultTimestampType().Equals(pt, nil) {
		return typ.Time
	}
	// dgo.Time is not a range
	panic(fmt.Errorf(`unable to create dgo time type from %s without loosing range constraint`, pt))
}

func fromTupleType(pt *types.TupleType) dgo.Type {
	if types.DefaultTupleType().Equals(pt, nil) {
		return typ.Tuple
	}
	sz := pt.Size()
	l := sz.Min()
	if l != sz.Max() {
		panic(fmt.Errorf(`unable to create dgo tuple type from %s without loosing size constraint`, pt))
	}
	tps := make([]dgo.Type, l)
	pts := pt.Types()
	pMax := len(pts) - 1
	lv := fromType(pts[pMax])
	for i := 0; i < int(l); i++ {
		var et dgo.Type
		if i >= pMax {
			et = lv
		} else {
			et = fromType(pts[i])
		}
		tps[i] = et
	}
	return newtype.Tuple(tps...)
}

func fromVariantType(pts ...px.Type) dgo.Type {
	tps := make([]dgo.Type, len(pts))
	for i := range pts {
		tps[i] = fromType(pts[i])
	}
	return newtype.AnyOf(tps...)
}
