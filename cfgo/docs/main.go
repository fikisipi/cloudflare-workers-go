//go:generate wget https://github.com/fikisipi/cloudflare-workers-go/releases/download/0.0.1/pkged.go

package main

import (
	"go/doc"
	"go/token"
	"go/ast"
	"go/parser"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"github.com/markbates/pkger"
	"io/ioutil"
	"os/exec"
	"runtime"
	"io/fs"
)
func main() {
	if inf, err := os.Stat("src"); err != nil || !inf.IsDir() {
		err = os.Mkdir("src", os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}
	summary, _ := os.Create("src/SUMMARY.md")
	summary.WriteString("# Summary\n\n- [cfgo](cfgo.md)\n")

	cnt := make(map[string]int)
	write := func(file *os.File, body string, a ...interface{}) {
		S := fmt.Sprintf(body, a...)
		file.WriteString(S)
		cnt[file.Name()] += len(strings.TrimSpace(S))
	}

	entries, _ := os.ReadDir("..")

	m := token.NewFileSet()
	files := make([]*ast.File, 0)
	for _, f := range entries {
		inf, _ := f.Info()
		m.AddFile("../" + f.Name(), m.Base(), int(inf.Size()))
	}
	myPkg, _ := parser.ParseDir(m, "..", nil, parser.ParseComments)
	for _, f := range myPkg["cfgo"].Files {
		files = append(files, f)
	}

	pkg, _ := doc.NewFromFiles(m, files, "github.com/fikisipi/cloudflare-workers-go/cfgo", doc.Mode(0))
	f, _ := os.Create("src/cfgo.md")
	write(f, pkg.Doc)


	fileTitles := make(map[string]string)
	for _, f := range files {
		fileName := (m.File(f.Pos()).Name())
		if strings.Contains(fileName, "_test") {
			for _, comm := range f.Comments {
				comStr := strings.TrimSpace(comm.Text())
				if strings.HasPrefix(comStr, "<DOCMAP>") {
					for _, line := range strings.Split(comStr, "\n") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							filename := strings.TrimLeft(parts[0], "1234567890 .")
							title := parts[1]
							fileTitles[filename] = title
						}
					}
				}
			}
		}
	}

	for _, entry := range entries {
		if entry.IsDir() { continue }

		title, ok := fileTitles[entry.Name()]
		if !ok { title = strings.TrimSuffix(entry.Name(), ".go"); }

		notHere := func(pos token.Pos) bool {
			return filepath.Base(m.File(pos).Name()) != entry.Name();
		}
		snippet := func(pos token.Pos, pos2 token.Pos) string {
			srcB, _ := os.ReadFile( m.File(pos).Name())
			srcStr := string(srcB)
			return srcStr[m.Position(pos).Offset : m.Position(pos2).Offset]
		}
		if strings.Contains(entry.Name(), "_test") {
			continue
		}
		mdFile, _ := os.Create("src/" + entry.Name() + ".md")
		mdFile.WriteString(fmt.Sprintf("# %s \n", title))

		for _, vr := range pkg.Vars {
			if notHere(vr.Decl.Pos()) { continue }
			varDecl := (snippet(vr.Decl.Pos(), vr.Decl.End()))
			write(mdFile, "%s\n```go\n%s\n```\n", vr.Doc, varDecl)
		}
		for _, function := range pkg.Funcs {
			if notHere(function.Decl.Pos()) { continue; }
			write(mdFile, "```go\n%s\n```\n", snippet(function.Decl.Pos(), function.Decl.End()))
			write(mdFile, "%s\n", function.Doc)
		}
		mdFile.WriteString("\n")
		for _, newF := range pkg.Types {
			if notHere(newF.Decl.Pos()) { continue; }
			for _, s := range newF.Decl.Specs {
				t := s.(*ast.TypeSpec)
				declName := t.Name.Name
				st, ok := t.Type.(*ast.StructType)
				if ok {
					write(mdFile, "## struct " + declName + "\n\n```go\ntype %s struct {\n", declName)
					for _, field := range st.Fields.List {
						write(mdFile, "  %s\n", snippet(field.Pos(), field.End()))
					}
					write(mdFile, "}\n```\n")
				} else {
					it, ok := t.Type.(*ast.InterfaceType)
					if !ok { continue }
					write(mdFile, "## interface %s\n```go\ntype %s interface {\n", declName, declName)
					for _, meth := range it.Methods.List {
						snip := (snippet(meth.Pos(), meth.End()))
						write(mdFile, "  %s\n", snip)
					}
					write(mdFile, "}\n```\n")
				}
				write(mdFile, "%s\n", newF.Doc)
			}
			for _, m := range newF.Methods {
				decl := m.Decl
				snip := (snippet(decl.Pos(), decl.End()))
				write(mdFile, fmt.Sprintf("```go\n%s\n```\n%s\n", snip, m.Doc))
				for _, e := range m.Examples {
					write(mdFile, "Example:\n```go\n%s\n```\n", snippet(e.Code.Pos(), e.Code.End()))
				}
			}
		}
		if cnt[mdFile.Name()] > 1 {
			write(summary, "   - [%s](%s.md)\n", title, entry.Name())
		}
		mdFile.Close()
	}
	win, _ := pkger.Open("/pkger/mdbook.exe")
	linux, _ := pkger.Open("/pkger/mdbook.bin")
	plat := win
	fmt.Println("Detected GOOS =", runtime.GOOS, "\n")
	if runtime.GOOS == "windows" {
		plat = win
	} else {
		plat = linux
	}
	platFile := filepath.Base(plat.Name())
	fileBytes, _ := ioutil.ReadAll(plat)
	ioutil.WriteFile(platFile, fileBytes, 0777)
	err := exec.Command("./" + platFile, "build", "-d", "src").Run()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Wrote docs to \"src/\".\n")
		fmt.Println("Directory contents:\n")
		filepath.WalkDir("./src", func(path string, d fs.DirEntry, err error) error {
			if strings.HasSuffix(path, "html") {
				fmt.Println(path)
			}
			return nil
		})
	}
	os.Remove(platFile)
}
