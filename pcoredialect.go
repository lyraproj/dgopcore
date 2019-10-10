package dgopcore

import (
	"github.com/lyraproj/dgo/dgo"
	"github.com/lyraproj/dgo/streamer"
	"github.com/lyraproj/dgo/vf"
	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/serialization"
)

// PcoreDialect returns the pcore dialect
func PcoreDialect() streamer.Dialect {
	return pcoreDialectSingleton
}

type pcoreDialect int

const pcoreDialectSingleton = pcoreDialect(0)

var typeKey = vf.String(serialization.PcoreTypeKey)
var valueKey = vf.String(serialization.PcoreValueKey)
var refKey = vf.String(serialization.PcoreRefKey)
var aliasType = vf.String(`Alias`)
var binaryType = vf.String(serialization.PcoreTypeBinary)
var sensitiveType = vf.String(serialization.PcoreTypeSensitive)
var mapType = vf.String(serialization.PcoreTypeHash)
var timeType = vf.String(`Timestamp`)

func (d pcoreDialect) TypeKey() dgo.String {
	return typeKey
}

func (d pcoreDialect) ValueKey() dgo.String {
	return valueKey
}

func (d pcoreDialect) RefKey() dgo.String {
	return refKey
}

func (d pcoreDialect) AliasTypeName() dgo.String {
	return aliasType
}

func (d pcoreDialect) BinaryTypeName() dgo.String {
	return binaryType
}

func (d pcoreDialect) MapTypeName() dgo.String {
	return mapType
}

func (d pcoreDialect) SensitiveTypeName() dgo.String {
	return sensitiveType
}

func (d pcoreDialect) TimeTypeName() dgo.String {
	return timeType
}

func (d pcoreDialect) ParseType(typeString dgo.String) (dt dgo.Type) {
	pcore.Do(func(c px.Context) {
		dt = fromType(c.ParseType(typeString.String()))
	})
	return
}
