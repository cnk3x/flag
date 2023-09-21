rm -rf *_test.go .tr* verify
sed -i 's/package pflag/package flag/gi; s|github.com/spf13/pflag|github.com/cnk3x/flag|g' *.go go.* *.md
sed -i 's|pflag: help requested||g' flag.go
go mod tidy
