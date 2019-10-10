package dgopcore_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/lyraproj/dgo/dgo"

	"github.com/lyraproj/dgo/typ"

	require "github.com/lyraproj/dgo/dgo_test"
	"github.com/lyraproj/dgo/dgopcore"
	"github.com/lyraproj/dgo/newtype"
	"github.com/lyraproj/dgo/vf"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
)

func ExampleToPcore_mixed() {
	v := vf.Map("a", 1, "b", vf.Values("c", 3.14, true))
	c := dgopcore.ToPcore(v)
	fmt.Println(c)
	// Output: {'a' => 1, 'b' => ['c', 3.14000, true]}
}

func ExampleToPcore_array() {
	v := vf.Strings("a", "b")
	c := dgopcore.ToPcore(v)
	fmt.Println(c)
	// Output: ['a', 'b']
}

func ExampleToPcore_map() {
	v := vf.Map("a", 1, "b", 2)
	c := dgopcore.ToPcore(v)
	fmt.Println(c)
	// Output: {'a' => 1, 'b' => 2}
}

func TestToPcore_binary(t *testing.T) {
	require.Equal(t, types.WrapBinary([]byte{1, 2}), dgopcore.ToPcore(vf.Value([]byte{1, 2})))
}

func TestToPcore_native(t *testing.T) {
	type X struct {
		A string
	}
	require.Equal(t, types.WrapRuntime(&X{A: `a`}), dgopcore.ToPcore(vf.Value(&X{A: `a`})))
}

func TestToPcore_regexp(t *testing.T) {
	rx := regexp.MustCompile(`abc`)
	require.Equal(t, types.WrapRegexp2(rx), dgopcore.ToPcore(vf.Value(rx)))
}

func TestToPcore_sensitive(t *testing.T) {
	require.Equal(t, types.WrapSensitive(types.WrapString(`secret`)), dgopcore.ToPcore(vf.Sensitive(`secret`)))
}

func TestToPcore_time(t *testing.T) {
	ts := time.Now()
	require.Equal(t, types.WrapTimestamp(ts), dgopcore.ToPcore(vf.Time(ts)))
}

// obscureValue is to get test coverage of unknown value types
type obscureValue int

func (o obscureValue) String() string {
	panic("implement me")
}

func (o obscureValue) Type() dgo.Type {
	return typ.Any
}

func (o obscureValue) Equals(other interface{}) bool {
	panic("implement me")
}

func (o obscureValue) HashCode() int {
	panic("implement me")
}

func TestToPcore_unknown(t *testing.T) {
	require.Panic(t, func() { dgopcore.ToPcore(obscureValue(0)) }, `unable to create a px.Value from a dgo any`)
}

func TestToPcore_nil(t *testing.T) {
	require.Equal(t, px.Undef, dgopcore.ToPcore(vf.Nil))
}

func TestToPcore_anyType(t *testing.T) {
	require.Equal(t, types.DefaultAnyType(), dgopcore.ToPcore(typ.Any))
}

func TestToPcore_anyOfType(t *testing.T) {
	require.Equal(t, types.DefaultVariantType(), dgopcore.ToPcore(typ.AnyOf))
}

func TestToPcore_arrayType(t *testing.T) {
	require.Equal(t, types.DefaultArrayType(), dgopcore.ToPcore(typ.Array))
	require.Equal(t, types.WrapValues([]px.Value{types.WrapString(`a`), types.WrapInteger(2)}).PType(), dgopcore.ToPcore(vf.Values(`a`, 2).Type()))
}

func TestToPcore_binaryType(t *testing.T) {
	require.Equal(t, types.DefaultBinaryType(), dgopcore.ToPcore(typ.Binary))
	require.Equal(t, types.WrapBinary([]byte{1, 2}).PType(), dgopcore.ToPcore(vf.Binary([]byte{1, 2}, false).Type()))
}

func TestToPcore_boolType(t *testing.T) {
	require.Equal(t, types.DefaultBooleanType(), dgopcore.ToPcore(typ.Boolean))
	require.Equal(t, types.NewBooleanType(false), dgopcore.ToPcore(typ.False))
	require.Equal(t, types.NewBooleanType(true), dgopcore.ToPcore(typ.True))
}

func TestToPcore_floatRange(t *testing.T) {
	require.Equal(t, types.DefaultFloatType(), dgopcore.ToPcore(typ.Float))
	require.Equal(t, types.NewFloatType(-3.3, 44.3), dgopcore.ToPcore(newtype.FloatRange(-3.3, 44.3, true)))
	require.Panic(t, func() { dgopcore.ToPcore(newtype.FloatRange(-3.3, 44.3, false)) }, `unable to create non inclusive pcore Float type`)
	require.Equal(t, types.NewFloatType(43.3, 43.3), dgopcore.ToPcore(vf.Float(43.3).Type()))
}

func TestToPcore_integerRange(t *testing.T) {
	require.Equal(t, types.DefaultIntegerType(), dgopcore.ToPcore(typ.Integer))
	require.Equal(t, types.NewIntegerType(-3, 44), dgopcore.ToPcore(newtype.IntegerRange(-3, 44, true)))
	require.Equal(t, types.NewIntegerType(-3, 43), dgopcore.ToPcore(newtype.IntegerRange(-3, 44, false)))
	require.Equal(t, types.NewIntegerType(43, 43), dgopcore.ToPcore(vf.Integer(43).Type()))
}

func TestToPcore_mapType(t *testing.T) {
	require.Equal(t, types.DefaultHashType(), dgopcore.ToPcore(typ.Map))
	require.Equal(t,
		types.WrapStringToStringMap(map[string]string{`a`: `one`, `b`: `two`}).PType(),
		dgopcore.ToPcore(vf.Map(`a`, `one`, `b`, `two`).Type()))
}

func TestToPcore_metaType(t *testing.T) {
	require.Equal(t, types.DefaultTypeType(), dgopcore.ToPcore(typ.Type))
	require.Equal(t, types.NewTypeType(types.DefaultTypeType()).PType(), dgopcore.ToPcore(typ.Type.Type()))
	require.Equal(t, types.NewTypeType(types.DefaultStringType()), dgopcore.ToPcore(typ.String.Type()))
}

func TestToPcore_nativeType(t *testing.T) {
	type X struct {
		A string
	}
	require.Equal(t, types.DefaultRuntimeType(), dgopcore.ToPcore(typ.Native))
	require.Equal(t, types.NewGoRuntimeType(&X{}), dgopcore.ToPcore(vf.Value(&X{}).Type()))
}

func TestToPcore_patternType(t *testing.T) {
	rx := regexp.MustCompile(`abc`)
	require.Equal(t, types.NewPatternType([]*types.RegexpType{types.NewRegexpTypeR(rx)}), dgopcore.ToPcore(newtype.Pattern(rx)))
}

func TestToPcore_regexpType(t *testing.T) {
	rx := regexp.MustCompile(`abc`)
	require.Equal(t, types.DefaultRegexpType(), dgopcore.ToPcore(typ.Regexp))
	require.Equal(t, types.NewRegexpTypeR(rx), dgopcore.ToPcore(vf.Regexp(rx).Type()))
}

func TestToPcore_sensitiveType(t *testing.T) {
	require.Equal(t, types.DefaultSensitiveType(), dgopcore.ToPcore(typ.Sensitive))
	require.Equal(t, types.NewSensitiveType(types.DefaultStringType()), dgopcore.ToPcore(newtype.Sensitive(typ.String)))
}

func TestToPcore_structType(t *testing.T) {
	require.Equal(t,
		types.NewStructType(
			[]*types.StructElement{
				types.NewStructElement(types.NewOptionalType(types.WrapString(`a`).PType()), types.DefaultFloatType()),
				types.NewStructElement(types.WrapString(`b`), types.DefaultIntegerType())}),
		dgopcore.ToPcore(newtype.StructFromMap(false, vf.Map(`a`, typ.Float, `b`, vf.Map(`type`, typ.Integer, `required`, true)))))

	require.Panic(t, func() {
		dgopcore.ToPcore(newtype.StructFromMap(true, vf.Map(`a`, typ.Float)))
	}, `unable to create pcore Struct that allows additional entries`)
}

func TestToPcore_timeType(t *testing.T) {
	require.Equal(t, types.DefaultTimestampType(), dgopcore.ToPcore(typ.Time))
	ts := time.Now()
	require.Equal(t, types.WrapTimestamp(ts).PType(), dgopcore.ToPcore(vf.Time(ts).Type()))
}

func TestToPcore_tupleType(t *testing.T) {
	require.Equal(t, types.DefaultTupleType(), dgopcore.ToPcore(typ.Tuple))
	require.Equal(t,
		types.NewTupleType([]px.Type{types.DefaultStringType(), types.DefaultIntegerType()}, types.NewIntegerType(2, 2)),
		dgopcore.ToPcore(newtype.Tuple(typ.String, typ.Integer)))
}

func TestToPcore_stringType(t *testing.T) {
	require.Equal(t, types.DefaultStringType(), dgopcore.ToPcore(typ.String))
	require.Equal(t, types.NewStringType(nil, `hello`), dgopcore.ToPcore(vf.String(`hello`).Type()))
	require.Equal(t, types.NewStringType(types.NewIntegerType(1, 10), ``), dgopcore.ToPcore(newtype.String(1, 10)))
	require.Equal(t, types.NewEnumType([]string{`a`}, true), dgopcore.ToPcore(newtype.CiString(`a`)))
}

func TestToPcore_invalidType(t *testing.T) {
	require.Panic(t, func() {
		dgopcore.ToPcore(newtype.AllOf(newtype.Pattern(regexp.MustCompile(`ab`)), newtype.String(20, 30)))
	}, `unable to create pcore Type from`)
}
