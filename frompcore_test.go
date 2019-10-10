package dgopcore_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/lyraproj/semver/semver"

	require "github.com/lyraproj/dgo/dgo_test"

	"github.com/lyraproj/dgo/newtype"

	"github.com/lyraproj/dgo/dgo"
	"github.com/lyraproj/dgo/typ"

	"github.com/lyraproj/pcore/types"

	"github.com/lyraproj/dgo/vf"

	"github.com/lyraproj/dgo/dgopcore"
	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
)

func ExampleFromPcore_mixed() {
	pcore.Do(func(c px.Context) {
		v := px.Wrap(c, map[string]interface{}{"a": 1, "b": []interface{}{"c", 3.14, true}})
		fmt.Println(dgopcore.FromPcore(v))
	})
	// Output: {"a":1,"b":["c",3.14,true]}
}

func requireDgo(t *testing.T, e, a interface{}) {
	t.Helper()
	ev := vf.Value(e)
	if !ev.Equals(a) {
		t.Errorf(`%s is not equal to %s`, ev, vf.Value(a))
	}
}

func TestFromPcore_values(t *testing.T) {
	values := []interface{}{
		true,
		false,
		nil,
		20,
		3.14,
		`hello`,
		[]byte{1, 2, 3},
		[]string{`a`, `b`, `c`},
		map[string]string{`a`: `first`, `b`: `second`},
		regexp.MustCompile(`abc`),
		time.Now(),
	}
	pcore.Do(func(c px.Context) {
		t.Helper()
		for _, v := range values {
			pv := px.Wrap(c, v)
			requireDgo(t, v, dgopcore.FromPcore(pv))
		}
	})
}

func TestFromPcore_sensitive(t *testing.T) {
	requireDgo(t, vf.Sensitive(`hello`), dgopcore.FromPcore(types.WrapSensitive(types.WrapString(`hello`))))
}

func TestFromPcore_defaultTypes(t *testing.T) {
	types := map[dgo.Type]px.Type{
		typ.Any:       types.DefaultAnyType(),
		typ.Array:     types.DefaultArrayType(),
		typ.Binary:    types.DefaultBinaryType(),
		typ.Boolean:   types.DefaultBooleanType(),
		typ.False:     types.NewBooleanType(false),
		typ.True:      types.NewBooleanType(true),
		typ.Float:     types.DefaultFloatType(),
		typ.Map:       types.DefaultHashType(),
		typ.Integer:   types.DefaultIntegerType(),
		typ.Regexp:    types.DefaultRegexpType(),
		typ.Sensitive: types.DefaultSensitiveType(),
		typ.String:    types.DefaultStringType(),
		typ.Time:      types.DefaultTimestampType(),
		typ.Tuple:     types.DefaultTupleType(),
		typ.Nil:       types.DefaultUndefType(),
	}
	for dt, pt := range types {
		requireDgo(t, dt, dgopcore.FromPcore(pt))
	}
}

func TestFromPcore_collectionType(t *testing.T) {
	require.Equal(t, newtype.AnyOf(typ.Array, typ.Map), dgopcore.FromPcore(types.DefaultCollectionType()))
}

func TestFromPcore_enumType(t *testing.T) {
	require.Equal(t, typ.String, dgopcore.FromPcore(types.DefaultEnumType()))
	require.Equal(t, vf.String(`a`).Type(), dgopcore.FromPcore(types.NewEnumType([]string{`a`}, false)))
	require.Equal(t, newtype.Enum(`a`, `b`), dgopcore.FromPcore(types.NewEnumType([]string{`a`, `b`}, false)))
	require.Equal(t, newtype.CiEnum(`a`, `b`), dgopcore.FromPcore(types.NewEnumType([]string{`a`, `b`}, true)))
}

func TestFromPcore_hashType(t *testing.T) {
	require.Equal(t, newtype.Map(typ.String, typ.Map, 3, 8),
		dgopcore.FromPcore(types.NewHashType(types.DefaultStringType(), types.DefaultHashType(), types.NewIntegerType(3, 8))))
}

func TestFromPcore_notUndefType(t *testing.T) {
	require.Equal(t, newtype.AllOf(newtype.Not(typ.Nil), typ.Any), dgopcore.FromPcore(types.DefaultNotUndefType()))
}

func TestFromPcore_numericType(t *testing.T) {
	require.Equal(t, newtype.AnyOf(typ.Float, typ.Integer), dgopcore.FromPcore(types.DefaultNumericType()))
}

func TestFromPcore_optionalType(t *testing.T) {
	require.Equal(t, typ.Any, dgopcore.FromPcore(types.DefaultOptionalType()))
	require.Equal(t, newtype.AnyOf(typ.Nil, typ.String), dgopcore.FromPcore(
		types.NewOptionalType(types.DefaultStringType())))
}

func TestFromPcore_patternType(t *testing.T) {
	require.Equal(t, typ.String, dgopcore.FromPcore(types.DefaultPatternType()))
	rx := regexp.MustCompile(`abc`)
	require.Equal(t, newtype.Pattern(rx), dgopcore.FromPcore(types.NewPatternType([]*types.RegexpType{
		types.NewRegexpTypeR(rx)})))
}

func TestFromPcore_regexpType(t *testing.T) {
	rx := regexp.MustCompile(`abc`)
	require.Equal(t, vf.Regexp(rx).Type(), dgopcore.FromPcore(types.NewRegexpTypeR(rx)))
}

func TestFromPcore_sensitiveType(t *testing.T) {
	requireDgo(t, vf.Sensitive(`hello`).Type(), dgopcore.FromPcore(types.WrapSensitive(types.WrapString(`hello`)).PType()))
}

func TestFromPcore_structType(t *testing.T) {
	require.Equal(t, typ.Map, dgopcore.FromPcore(types.DefaultStructType()))
	require.Equal(t,
		newtype.StructFromMap(false, vf.Map(`a`, typ.Float, `b`, vf.Map(`type`, typ.Integer, `required`, true))),
		dgopcore.FromPcore(types.NewStructType(
			[]*types.StructElement{
				types.NewStructElement(types.NewOptionalType(types.WrapString(`a`).PType()), types.DefaultFloatType()),
				types.NewStructElement(types.WrapString(`b`), types.DefaultIntegerType())})))
}

func TestFromPcore_timestampType(t *testing.T) {
	require.Panic(t, func() {
		ts := time.Now()
		dgopcore.FromPcore(types.NewTimestampType(ts, ts.Add(1*time.Second)))
	}, `unable to create dgo time type from Timestamp\[.*without loosing range constraint`)
}

func TestFromPcore_tupleType(t *testing.T) {
	require.Equal(t, newtype.Tuple(typ.Float, typ.Integer), dgopcore.FromPcore(types.NewTupleType(
		[]px.Type{types.DefaultFloatType(), types.DefaultIntegerType()}, nil)))
	require.Panic(t, func() {
		dgopcore.FromPcore(types.NewTupleType(
			[]px.Type{types.DefaultFloatType(), types.DefaultIntegerType()}, types.NewIntegerType(2, 4)))
	}, `unable to create dgo tuple type from Tuple\[Float, Integer, 2, 4\] without loosing size constraint`)
}

func TestFromPcore_typeType(t *testing.T) {
	require.Equal(t, typ.Type, dgopcore.FromPcore(types.DefaultTypeType()))
	require.Equal(t, typ.String.Type(), dgopcore.FromPcore(types.DefaultStringType().PType()))
}

func TestFromPcore_variantType(t *testing.T) {
	require.Equal(t, newtype.Not(typ.Any), dgopcore.FromPcore(types.DefaultVariantType()))
	require.Equal(t,
		newtype.AnyOf(typ.String, typ.Integer),
		dgopcore.FromPcore(types.NewVariantType(types.DefaultStringType(), types.DefaultIntegerType())))
}

func TestFromPcore_invalidValue(t *testing.T) {
	require.Panic(t, func() {
		dgopcore.FromPcore(types.WrapSemVer(semver.MustParseVersion(`1.0.0`)))
	}, `unable to create a dgo.Value from a pcore SemVer`)
}

func TestFromPcore_invalidType(t *testing.T) {
	require.Panic(t, func() {
		dgopcore.FromPcore(types.NewSemVerType(semver.MustParseVersionRange(`1.x`)))
	}, `unable to create a dgo.Type from a pcore SemVer`)
}
