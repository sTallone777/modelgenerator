package generate

import (
	"fmt"
	"io"
	"modelgenerator/conf"
	"modelgenerator/db"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/infobloxopen/atlas-app-toolkit/util"
)

var wg sync.WaitGroup

func Generate(tableNames ...string) {
	tableNamesStr := ""
	if len(tableNames) > 0 {
		tableNamesStr = "'" + strings.Join(tableNames, "' ,'") + "'"
	}
	tables := getTables(tableNamesStr)
	wg.Add(len(tables))
	for _, table := range tables {
		fields := getFields(table.Name)
		go generateModel(table, fields)
	}
	wg.Wait()
	fmt.Println("Program completed.")
}

// Get tables count
func getTables(tableNames string) []Table {

	query := db.Instance.Debug()
	var tables []Table

	if tableNames == "" {
		query.Raw("SELECT name FROM sysobjects WHERE xtype = 'u' AND name != 'sysdiagrams' ORDER BY name;").Find(&tables)
	} else {
		query.Raw("SELECT name FROM sysobjects WHERE xtype = 'u' AND name != 'sysdiagrams' AND name in (" + tableNames + ") ORDER BY name;").Find(&tables)
	}

	return tables
}

// Get field properties
func getFields(tableName string) []Field {

	query := db.Instance.Debug()
	var fields []Field

	query.Raw("SELECT " +
		"    B.NAME field_name, " +
		"    C.NAME field_type, " +
		"    CASE WHEN B.ISNULLABLE = 0 THEN 'no' ELSE 'yes' END field_isnullable," +
		"    B.PREC field_size, " +
		"    CONVERT(VARCHAR(20), ISNULL(B.SCALE, '')) field_decimal, " +
		"    CASE WHEN NOT F.ID IS NULL THEN 'yes' ELSE 'no' END field_isprimarykey, " +
		"    CASE WHEN COLUMNPROPERTY(B.ID, B.NAME, 'ISIDENTITY') = 1 THEN 'yes' ELSE 'no' END AS field_isincrement, " +
		"    CONVERT(VARCHAR(1000), ISNULL(G.VALUE, '')) field_comment " +
		"FROM " +
		"SYSOBJECTS A " +
		"    INNER JOIN SYSCOLUMNS B ON A.ID = B.ID " +
		"    INNER JOIN SYSTYPES C ON B.XTYPE = C.XUSERTYPE " +
		"    LEFT JOIN SYSOBJECTS D ON B.ID = D.PARENT_OBJ " +
		"AND D.XTYPE = 'PK' " +
		"    LEFT JOIN SYSINDEXES E ON B.ID = E.ID " +
		"AND D.NAME = E.NAME " +
		"    LEFT JOIN SYSINDEXKEYS F ON B.ID = F.ID " +
		"AND B.COLID = F.COLID " +
		"AND E.INDID = F.INDID " +
		"    LEFT JOIN SYS.EXTENDED_PROPERTIES G ON B.ID = G.MAJOR_ID " +
		"AND B.COLID = G.MINOR_ID " +
		"WHERE " +
		"    A.XTYPE = 'U' " +
		"AND OBJECT_NAME(B.ID)= '" + tableName + "';").Find(&fields)
	return fields
}

func generateModel(table Table, fields []Field) {
	content := "package models\n\n"
	content += "import \"time\"\n\n"
	content += "type " + util.Camel(table.Name) + " struct {\n"
	for _, field := range fields {
		fieldName := util.Camel(field.Field)
		fieldJson := getFieldJson([]rune(field.Field))
		fieldGorm := getFieldGorm(field)
		fieldType := getFiledType(field)
		fieldComment := getFieldComment(field)
		content += "	" + fieldName + " " + fieldType + " `" + fieldGorm + fieldJson + fieldComment + "`" + "\n"
	}
	content += "}\n"

	content += "func (entity *" + util.Camel(table.Name) + ") TableName() string {\n"
	content += "	" + `return "` + table.Name + `"`
	content += "\n}"

	filename := conf.ModelPath + table.Name + ".go"
	var f *os.File
	var err error
	if checkFileIsExist(filename) {
		fmt.Println(table.Name + " was already exists, please generate again...")
		wg.Done()
		return
	}

	f, err = os.Create(filename)
	if err != nil {
		panic(err)
	}

	_, err = io.WriteString(f, content)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(util.Camel(table.Name) + " generated...")
	}

	defer f.Close()
	wg.Done()
}

// Make json field name
func getFieldJson(field []rune) string {
	for i, v := range field {
		field[i] = unicode.ToLower(v)
	}
	return ` json:"` + string(field) + `"`
}

func getFieldGorm(field Field) string {
	fieldContext := `gorm:"column:` + field.Field + `;type:` + field.Type + `(` + field.Size

	dec, _ := strconv.Atoi(field.Decimal)
	if dec > 0 {
		fieldContext = fieldContext + `,` + field.Decimal + `)`
	} else {
		fieldContext = fieldContext + `)`
	}
	if field.Key == "yes" {
		fieldContext = fieldContext + `;primaryKey`
	}
	if field.Extra == "yes" {
		fieldContext = fieldContext + `;autoIncrement`
	}
	if field.Nullable == "no" {
		fieldContext = fieldContext + `;not null`
	}
	return fieldContext + `"`
}

func getFieldComment(field Field) string {
	if len(field.Comment) > 0 {
		return ` comment:"` + field.Comment + `"`
	}
	return ""
}

func checkFileIsExist(filename string) bool {
	var exist = true

	if _, err := os.Stat(conf.ModelPath); os.IsNotExist(err) {
		if err := os.Mkdir(conf.ModelPath, os.ModePerm); err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		}
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func getFiledType(field Field) string {
	switch field.Type {
	case "int":
		return conf.Pointer + "int32"
	case "integer":
		return conf.Pointer + "int32"
	case "mediumint":
		return conf.Pointer + "int32"
	case "bit":
		return conf.Pointer + "int32"
	case "year":
		return conf.Pointer + "int32"
	case "smallint":
		return conf.Pointer + "int16"
	case "tinyint":
		return conf.Pointer + "int8"
	case "bigint":
		return conf.Pointer + "int64"
	case "decimal":
		return conf.Pointer + "float64"
	case "double":
		return conf.Pointer + "float32"
	case "float":
		return conf.Pointer + "float32"
	case "real":
		return conf.Pointer + "float32"
	case "numeric":
		return conf.Pointer + "float32"
	case "smalldatetime":
		return conf.Pointer + "time.Time"
	case "timestamp":
		return conf.Pointer + "time.Time"
	case "datetime":
		return conf.Pointer + "time.Time"
	case "time":
		return conf.Pointer + "time.Time"
	case "date":
		return conf.Pointer + "time.Time"
	default:
		return conf.Pointer + "string"
	}
}
