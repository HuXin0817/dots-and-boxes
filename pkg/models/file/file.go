package file

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func GenByteArray(filePath string) error {
	VarName := filepath.Base(filePath)
	VarName = strings.Replace(VarName, ".", "_", -1)
	VarName = strings.Replace(VarName, "-", "_", -1)
	VarName = strings.Replace(VarName, " ", "_", -1)
	VarName = CapitalizeFirst(VarName)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`package %s

var %s = []byte{`, filepath.Dir(filePath), VarName))

	log.Println("Gen", filePath+".go ...")
	f, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	for i, b := range f {
		builder.WriteString(fmt.Sprint(b))
		if i < len(f)-1 {
			builder.WriteString(", ")
		}
	}

	builder.WriteString("}\n")

	goVarFile, err := os.Create(filePath + ".go")
	if err != nil {
		return err
	}

	if _, err := goVarFile.WriteString(builder.String()); err != nil {
		return err
	}

	return nil
}
