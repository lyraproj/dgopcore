package dgopcore_test

import (
	"testing"

	"github.com/lyraproj/dgo/dgopcore"
	"github.com/lyraproj/dgo/vf"

	"github.com/lyraproj/dgo/newtype"
	"github.com/lyraproj/dgo/typ"

	"github.com/lyraproj/dgo/streamer"

	require "github.com/lyraproj/dgo/dgo_test"

	"github.com/lyraproj/dgo/dgo"
)

func TestDgoDialect_names(t *testing.T) {
	for k, v := range map[string]func(streamer.Dialect) dgo.String{
		`Alias`:     streamer.Dialect.AliasTypeName,
		`Binary`:    streamer.Dialect.BinaryTypeName,
		`Hash`:      streamer.Dialect.MapTypeName,
		`Sensitive`: streamer.Dialect.SensitiveTypeName,
		`Timestamp`: streamer.Dialect.TimeTypeName,
		`__pref`:    streamer.Dialect.RefKey,
		`__ptype`:   streamer.Dialect.TypeKey,
		`__pvalue`:  streamer.Dialect.ValueKey,
	} {
		require.Equal(t, k, v(dgopcore.PcoreDialect()))
	}
}

func TestDgoDialect_ParseType(t *testing.T) {
	require.Equal(t, newtype.Array(typ.String, 3, 8), dgopcore.PcoreDialect().ParseType(vf.String(`Array[String,3,8]`)))
}
