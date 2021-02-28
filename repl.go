package main

import (
	"bufio"
	"fmt"
	"github.com/gogn"
	"os"
	"strings"
)

func main() {
	mb := NewMemoryBackend()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to gogn SQL")

	for {
		fmt.Print("# ")
		text, err := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		ast, err := Parse(text)
		if err != nil {
			panic(err)
		}

		for _, stmt := range ast.Statements {
			switch stmt.Kind {
			case CreateTableKind:
				err = mb.CreateTable((stmt.CreateTableStatement))
				if err != nil {
					panic(err)
				}
				fmt.Println("ok")
			case InsertKind:
				err = mb.Insert(stmt.InsertStatement)
				if err != nil {
					panic(err)
				}
				fmt.Println("ok")
			case SelectKind:
				results, err := mb.Select(stmt.SelectStatement)
				if err != nil {
					panic(err)
				}

				for _, col := range results.Columns {
					fmt.Printf("| %s ", col.Name)
				}
				fmt.Println("|")

				for i := 0; i < 20; i++ {
					fmt.Printf("=")
				}
				fmt.Println()

				for _, result := range results.Rows {
					fmt.Printf("|")

					for i, cell := range result {
						typ := results.Columns[i].Type
						s := ""
						switch typ {
						case IntType:
							s = fmt.Sprintf("%d", cell.AsInt())
						case TextType:
							s = cell.AsText()
						}

						fmt.Printf(" %s | ", s)
					}
					fmt.Println()
				}
				fmt.Println("ok")
			}
		}
	}
}
