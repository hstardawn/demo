package main

import (
	"github.com/spf13/cobra"
	"github.com/zjutjh/mygo/foundation/command"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"app/dao/model"
	"app/register"
)

var tables = []string{
	"user",
	"block",
	"comment",
	"post",
}

func main() {
	command.Execute(
		register.Boot,    // 应用引导注册器
		register.Command, // 应用命令注册器
		func(cmd *cobra.Command, args []string) error {
			return nil
		},
	)

	g := gen.NewGenerator(gen.Config{
		OutPath: "./dao/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})
	g.UseDB(ndb.Pick())

	m := map[string]func(detailType gorm.ColumnType) (fieldType string){
		"tinyint": func(detailType gorm.ColumnType) (fieldType string) {
			return "int8"
		},
	}
	g.WithDataTypeMap(m)

	for _, table := range tables {
		tableName := g.GenerateModel(table)
		g.ApplyBasic(tableName)
	}
	for _, table := range tables {
		tableName := g.GenerateModel(table,
			gen.FieldIgnore("id", "ctime", "utime"),
			gen.FieldRelateModel(field.BelongsTo, "", model.BaseModel{},
				&field.RelateConfig{
					JSONTag: "-",
				}),
		)
		g.ApplyBasic(tableName)
	}

	g.Execute()
}
