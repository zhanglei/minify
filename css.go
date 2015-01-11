package minify

/* TODO: (non-exhaustive)
- remove space before !important
- collapse margin/padding/border/background/list/etc. definitions into one
- remove duplicate selector blocks
- remove quotes within url()?
*/

/*
Uses http://www.w3.org/TR/2010/PR-css3-color-20101028/ for colors
*/

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/tdewolff/css"
)

var epsilon = 0.00001

var shortenColorHex = map[string][]byte{
	"#000080": []byte("navy"),
	"#008000": []byte("green"),
	"#008080": []byte("teal"),
	"#4b0082": []byte("indigo"),
	"#800000": []byte("maroon"),
	"#800080": []byte("purple"),
	"#808000": []byte("olive"),
	"#808080": []byte("gray"),
	"#a0522d": []byte("sienna"),
	"#a52a2a": []byte("brown"),
	"#c0c0c0": []byte("silver"),
	"#cd853f": []byte("peru"),
	"#d2b48c": []byte("tan"),
	"#da70d6": []byte("orchid"),
	"#dda0dd": []byte("plum"),
	"#ee82ee": []byte("violet"),
	"#f0e68c": []byte("khaki"),
	"#f0ffff": []byte("azure"),
	"#f5deb3": []byte("wheat"),
	"#f5f5dc": []byte("beige"),
	"#fa8072": []byte("salmon"),
	"#faf0e6": []byte("linen"),
	"#ff6347": []byte("tomato"),
	"#ff7f50": []byte("coral"),
	"#ffa500": []byte("orange"),
	"#ffc0cb": []byte("pink"),
	"#ffd700": []byte("gold"),
	"#ffe4c4": []byte("bisque"),
	"#fffafa": []byte("snow"),
	"#fffff0": []byte("ivory"),
	"#ff0000": []byte("red"),
	"#f00":    []byte("red"),
}

var shortenColorName = map[string][]byte{
	"black":                []byte("#000"),
	"darkblue":             []byte("#00008b"),
	"mediumblue":           []byte("#0000cd"),
	"darkgreen":            []byte("#006400"),
	"darkcyan":             []byte("#008b8b"),
	"deepskyblue":          []byte("#00bfff"),
	"darkturquoise":        []byte("#00ced1"),
	"mediumspringgreen":    []byte("#00fa9a"),
	"springgreen":          []byte("#00ff7f"),
	"midnightblue":         []byte("#191970"),
	"dodgerblue":           []byte("#1e90ff"),
	"lightseagreen":        []byte("#20b2aa"),
	"forestgreen":          []byte("#228b22"),
	"seagreen":             []byte("#2e8b57"),
	"darkslategray":        []byte("#2f4f4f"),
	"limegreen":            []byte("#32cd32"),
	"mediumseagreen":       []byte("#3cb371"),
	"turquoise":            []byte("#40e0d0"),
	"royalblue":            []byte("#4169e1"),
	"steelblue":            []byte("#4682b4"),
	"darkslateblue":        []byte("#483d8b"),
	"mediumturquoise":      []byte("#48d1cc"),
	"darkolivegreen":       []byte("#556b2f"),
	"cadetblue":            []byte("#5f9ea0"),
	"cornflowerblue":       []byte("#6495ed"),
	"mediumaquamarine":     []byte("#66cdaa"),
	"slateblue":            []byte("#6a5acd"),
	"olivedrab":            []byte("#6b8e23"),
	"slategray":            []byte("#708090"),
	"lightslateblue":       []byte("#789"),
	"mediumslateblue":      []byte("#7b68ee"),
	"lawngreen":            []byte("#7cfc00"),
	"chartreuse":           []byte("#7fff00"),
	"aquamarine":           []byte("#7fffd4"),
	"lightskyblue":         []byte("#87cefa"),
	"blueviolet":           []byte("#8a2be2"),
	"darkmagenta":          []byte("#8b008b"),
	"saddlebrown":          []byte("#8b4513"),
	"darkseagreen":         []byte("#8fbc8f"),
	"lightgreen":           []byte("#90ee90"),
	"mediumpurple":         []byte("#9370db"),
	"darkviolet":           []byte("#9400d3"),
	"palegreen":            []byte("#98fb98"),
	"darkorchid":           []byte("#9932cc"),
	"yellowgreen":          []byte("#9acd32"),
	"darkgray":             []byte("#a9a9a9"),
	"lightblue":            []byte("#add8e6"),
	"greenyellow":          []byte("#adff2f"),
	"paleturquoise":        []byte("#afeeee"),
	"lightsteelblue":       []byte("#b0c4de"),
	"powderblue":           []byte("#b0e0e6"),
	"firebrick":            []byte("#b22222"),
	"darkgoldenrod":        []byte("#b8860b"),
	"mediumorchid":         []byte("#ba55d3"),
	"rosybrown":            []byte("#bc8f8f"),
	"darkkhaki":            []byte("#bdb76b"),
	"mediumvioletred":      []byte("#c71585"),
	"indianred":            []byte("#cd5c5c"),
	"chocolate":            []byte("#d2691e"),
	"lightgray":            []byte("#d3d3d3"),
	"goldenrod":            []byte("#daa520"),
	"palevioletred":        []byte("#db7093"),
	"gainsboro":            []byte("#dcdcdc"),
	"burlywood":            []byte("#deb887"),
	"lightcyan":            []byte("#e0ffff"),
	"lavender":             []byte("#e6e6fa"),
	"darksalmon":           []byte("#e9967a"),
	"palegoldenrod":        []byte("#eee8aa"),
	"lightcoral":           []byte("#f08080"),
	"aliceblue":            []byte("#f0f8ff"),
	"honeydew":             []byte("#f0fff0"),
	"sandybrown":           []byte("#f4a460"),
	"whitesmoke":           []byte("#f5f5f5"),
	"mintcream":            []byte("#f5fffa"),
	"ghostwhite":           []byte("#f8f8ff"),
	"antiquewhite":         []byte("#faebd7"),
	"lightgoldenrodyellow": []byte("#fafad2"),
	"fuchsia":              []byte("#f0f"),
	"magenta":              []byte("#f0f"),
	"deeppink":             []byte("#ff1493"),
	"orangered":            []byte("#ff4500"),
	"darkorange":           []byte("#ff8c00"),
	"lightsalmon":          []byte("#ffa07a"),
	"lightpink":            []byte("#ffb6c1"),
	"peachpuff":            []byte("#ffdab9"),
	"navajowhite":          []byte("#ffdead"),
	"moccasin":             []byte("#ffe4b5"),
	"mistyrose":            []byte("#ffe4e1"),
	"blanchedalmond":       []byte("#ffebcd"),
	"papayawhip":           []byte("#ffefd5"),
	"lavenderblush":        []byte("#fff0f5"),
	"seashell":             []byte("#fff5ee"),
	"cornsilk":             []byte("#fff8dc"),
	"lemonchiffon":         []byte("#fffacd"),
	"floralwhite":          []byte("#fffaf0"),
	"yellow":               []byte("#ff0"),
	"lightyellow":          []byte("#ffffe0"),
	"white":                []byte("#fff"),
}

// CSS minifies CSS files, it reads from r and writes to w.
func (m Minifier) CSS(w io.Writer, r io.Reader) error {
	stylesheet, err := css.Parse(r)
	if err != nil {
		return err
	}
	shortenNodes(stylesheet.Nodes)
	return writeNodes(w, stylesheet.Nodes)
}

////////////////////////////////////////////////////////////////

func shortenNodes(nodes []css.Node) {
	inHeader := true
	for i, n := range nodes {
		if _, ok := n.(*css.NodeAtRule); inHeader && !ok {
			inHeader = false
		}
		switch m := n.(type) {
		case *css.NodeDeclaration:
			shortenDecl(m)
		case *css.NodeRuleset:
			for _, selGroup := range m.SelGroups {
				for _, sel := range selGroup.Selectors {
					shortenSelector(sel)
				}
			}
			for _, decl := range m.Decls {
				shortenDecl(decl)
			}
		case *css.NodeAtRule:
			if !inHeader && (bytes.Equal(m.At.Data, []byte("@charset")) || bytes.Equal(m.At.Data, []byte("@import"))) {
				nodes[i] = css.NewToken(css.DelimToken, []byte(""))
			}
			if m.Block != nil {
				shortenNodes(m.Block.Nodes)
			}
		case *css.NodeBlock:
			shortenNodes(m.Nodes)
		}
	}
}

func shortenSelector(sel *css.NodeSelector) {
	class := false
	for i, n := range sel.Nodes {
		switch m := n.(type) {
		case *css.NodeToken:
			if m.TokenType == css.DelimToken && m.Data[0] == '.' {
				class = true
			} else if m.TokenType == css.IdentToken {
				if !class {
					m.Data = bytes.ToLower(m.Data)
				}
				class = false
			}
		case *css.NodeAttributeSelector:
			for j, val := range m.Vals {
				if val.TokenType == css.StringToken {
					s := val.Data[1 : len(val.Data)-1]
					if css.IsIdent([]byte(s)) {
						m.Vals[j] = css.NewToken(css.IdentToken, s)
					}
				}
			}
			if (bytes.Equal(m.Key.Data, []byte("id")) || bytes.Equal(m.Key.Data, []byte("class"))) && m.Op != nil && bytes.Equal(m.Op.Data, []byte("=")) && len(m.Vals) == 1 && css.IsIdent(m.Vals[0].Data) {
				if bytes.Equal(m.Key.Data, []byte("id")) {
					sel.Nodes[i] = css.NewToken(css.HashToken, append([]byte("#"), m.Vals[0].Data...))
				} else {
					sel.Nodes[i] = css.NewToken(css.DelimToken, []byte("."))
					sel.Nodes = append(append(sel.Nodes[:i+1], css.NewToken(css.IdentToken, m.Vals[0].Data)), sel.Nodes[i+1:]...)
					class = true
				}
			}
		}
	}
}

func shortenDecl(decl *css.NodeDeclaration) {
	// shorten zeros
	progid := false
	for i, n := range decl.Vals {
		switch m := n.(type) {
		case *css.NodeToken:
			if !progid {
				if i == 0 && bytes.Equal(m.Data, []byte("progid")) {
					progid = true
					continue
				}
				decl.Vals[i] = shortenToken(m)
			}
		case *css.NodeFunction:
			if !progid {
				m.Func.Data = bytes.ToLower(m.Func.Data)
			}
			for j, arg := range m.Args {
				m.Args[j].Val = shortenToken(arg.Val)
			}
		}
	}

	decl.Prop.Data = bytes.ToLower(decl.Prop.Data)
	prop := decl.Prop.Data
	if bytes.Equal(prop, []byte("margin")) || bytes.Equal(prop, []byte("padding")) {
		tokens := make([]*css.NodeToken, 0, 4)
		for _, n := range decl.Vals {
			if m, ok := n.(*css.NodeToken); ok {
				tokens = append(tokens, m)
			} else {
				tokens = []*css.NodeToken{}
				break
			}
		}
		if len(tokens) == 2 {
			if bytes.Equal(tokens[0].Data, tokens[1].Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			}
		} else if len(tokens) == 3 {
			if bytes.Equal(tokens[0].Data, tokens[1].Data) && bytes.Equal(tokens[0].Data, tokens[2].Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(tokens[0].Data, tokens[2].Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			}
		} else if len(tokens) == 4 {
			if bytes.Equal(tokens[0].Data, tokens[1].Data) && bytes.Equal(tokens[0].Data, tokens[2].Data) && bytes.Equal(tokens[0].Data, tokens[3].Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(tokens[0].Data, tokens[2].Data) && bytes.Equal(tokens[1].Data, tokens[3].Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			} else if bytes.Equal(tokens[1].Data, tokens[3].Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1], decl.Vals[2]}
			}
		}
	} else if bytes.HasPrefix(prop, []byte("font")) && (len(prop) == len("font") || bytes.Equal(prop, []byte("font-family"))) {
		for i, n := range decl.Vals {
			if m, ok := n.(*css.NodeToken); ok {
				if m.TokenType == css.IdentToken {
					if bytes.Equal(m.Data, []byte("normal")) && bytes.Equal(prop, []byte("font-weight")) { // normal could also be specified for font-variant, not just font-weight
						decl.Vals[i] = css.NewToken(css.NumberToken, []byte("400"))
					} else if bytes.Equal(m.Data, []byte("bold")) {
						decl.Vals[i] = css.NewToken(css.NumberToken, []byte("700"))
					}
				} else if m.TokenType == css.StringToken {
					m.Data = bytes.ToLower(m.Data)
					s := m.Data[1 : len(m.Data)-1]
					unquote := true
					for _, fontName := range bytes.Split(s, []byte(" ")) {
						// if len is zero, it contains two consecutive spaces
						if len(fontName) == 0 || !css.IsIdent(fontName) || bytes.Equal(fontName, []byte("inherit")) || bytes.Equal(fontName, []byte("serif")) || bytes.Equal(fontName, []byte("sans-serif")) || bytes.Equal(fontName, []byte("monospace")) ||
						   bytes.Equal(fontName, []byte("fantasy")) || bytes.Equal(fontName, []byte("cursive")) || bytes.Equal(fontName, []byte("initial")) || bytes.Equal(fontName, []byte("default")) {
							unquote = false
							break
						}
					}
					if unquote {
						m.Data = s
					}
				}
			}
		}
	} else if len(decl.Vals) == 7 && bytes.Equal(prop, []byte("filter")) {
		if n, ok := decl.Vals[6].(*css.NodeFunction); ok && bytes.Equal(n.Func.Data, []byte("Alpha(")) {
			tokens := []byte{}
			for _, val := range decl.Vals[:len(decl.Vals)-1] {
				if m, ok := val.(*css.NodeToken); ok {
					tokens = append(tokens, m.Data...)
				} else {
					tokens = []byte{}
					break
				}
			}
			f := decl.Vals[6].(*css.NodeFunction)
			if bytes.Equal(tokens, []byte("progid:DXImageTransform.Microsoft.")) && len(f.Args) == 1 && bytes.Equal(f.Args[0].Key.Data, []byte("Opacity")) {
				newF := css.NewFunction(css.NewToken(css.FunctionToken, []byte("alpha(")))
				newF.Args = f.Args
				newF.Args[0].Key.Data = bytes.ToLower(newF.Args[0].Key.Data)
				decl.Vals = []css.Node{newF}
			}
		}
	} else if len(decl.Vals) == 1 && bytes.Equal(prop, []byte("-ms-filter")) {
			if n, ok := decl.Vals[0].(*css.NodeToken); ok {
				alpha := []byte("progid:DXImageTransform.Microsoft.Alpha(Opacity=")
				if n.TokenType == css.StringToken && bytes.HasPrefix(n.Data[1:len(n.Data)-1], alpha) {
					n.Data = append(append([]byte{n.Data[0]}, []byte("alpha(opacity=")...), n.Data[1+len(alpha):]...)
				}
			}
	} else {
		if len(decl.Vals) == 1 && (bytes.Equal(prop, []byte("outline")) || bytes.Equal(prop, []byte("background")) ||
			bytes.HasPrefix(prop, []byte("border")) && (len(prop) == len("border") || bytes.Equal(prop, []byte("border-top")) || bytes.Equal(prop, []byte("border-right")) || bytes.Equal(prop, []byte("border-bottom")) || bytes.Equal(prop, []byte("border-left")))) {
			if n, ok := decl.Vals[0].(*css.NodeToken); ok && bytes.Equal(bytes.ToLower(n.Data), []byte("none")) {
				decl.Vals[0] = css.NewToken(css.NumberToken, []byte("0"))
			}
		}

		for i, n := range decl.Vals {
			switch m := n.(type) {
			case *css.NodeToken:
				if m.TokenType == css.URLToken {
					s := m.Data[4 : len(m.Data)-1]
					if s[0] == '"' || s[0] == '\'' {
						if css.IsUrlUnquoted([]byte(s[1 : len(s)-1])) {
							s = s[1 : len(s)-1]
						}
					}
					m.Data = append(append([]byte("url("), s...), ')')
				}
			case *css.NodeFunction:
				if bytes.Equal(m.Func.Data, []byte("rgba(")) && len(m.Args) == 4 {
					d, _ := strconv.ParseFloat(string(m.Args[3].Val.Data), 32)
					if math.Abs(d-1.0) < epsilon {
						m.Func = css.NewToken(css.FunctionToken, []byte("rgb("))
						m.Args = m.Args[:len(m.Args)-1]
					}
				}
				if bytes.Equal(m.Func.Data, []byte("rgb(")) && len(m.Args) == 3 {
					var err error
					rgb := make([]byte, 3)
					for j := 0; j < 3; j++ {
						v := m.Args[j].Val
						if v.TokenType == css.NumberToken {
							var d int64
							d, err = strconv.ParseInt(string(v.Data), 10, 32)
							if d < 0 {
								d = 0
							} else if d > 255 {
								d = 255
							}
							rgb[j] = byte(d)
						} else if v.TokenType == css.PercentageToken {
							var d float64
							d, err = strconv.ParseFloat(string(v.Data[:len(v.Data)-1]), 32)
							if d < 0.0 {
								d = 0.0
							} else if d > 100.0 {
								d = 100.0
							}
							rgb[j] = byte((d / 100.0 * 255.0) + 0.5)
						} else {
							err = errors.New("")
							break
						}
					}
					if err == nil {
						valHex := make([]byte, 6)
						hex.Encode(valHex, rgb)
						val := append([]byte("#"), bytes.ToLower(valHex)...)
						if s, ok := shortenColorHex[string(val)]; ok {
							decl.Vals[i] = css.NewToken(css.IdentToken, s)
						} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
							decl.Vals[i] = css.NewToken(css.HashToken, append([]byte("#"), val[1], val[3], val[5]))
						} else {
							decl.Vals[i] = css.NewToken(css.HashToken, val)
						}
					}
				}
			}
		}
	}
}

func shortenToken(token *css.NodeToken) *css.NodeToken {
	if token.TokenType == css.NumberToken || token.TokenType == css.DimensionToken || token.TokenType == css.PercentageToken {
		token.Data = bytes.ToLower(token.Data)
		if len(token.Data) > 0 && token.Data[0] == '+' {
			token.Data = token.Data[1:]
		}

		num, dim := css.SplitNumberToken(token.Data)
		f, err := strconv.ParseFloat(string(num), 64)
		if err != nil {
			return token
		}
		if math.Abs(f) < epsilon {
			token.Data = []byte("0")
		} else {
			if len(num) > 0 && num[0] == '-' {
				num = append([]byte{'-'}, bytes.TrimLeft(num[1:], "0")...)
			} else {
				num = bytes.TrimLeft(num, "0")
			}
			if bytes.Index(num, []byte(".")) != -1 {
				num = bytes.TrimRight(num, "0")
				if num[len(num)-1] == '.' {
					num = num[:len(num)-1]
				}
			}
			token.Data = append(num, dim...)
		}
	} else if token.TokenType == css.IdentToken {
		token.Data = bytes.ToLower(token.Data)
		if h, ok := shortenColorName[string(token.Data)]; ok {
			token = css.NewToken(css.HashToken, h)
		}
	} else if token.TokenType == css.HashToken {
		val := bytes.ToLower(token.Data)
		if i, ok := shortenColorHex[string(val)]; ok {
			token = css.NewToken(css.IdentToken, i)
		} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
			token = css.NewToken(css.HashToken, append([]byte("#"), bytes.ToLower(append([]byte{val[1]}, val[3], val[5]))...))
		} else {
			token.Data = bytes.ToLower(token.Data)
		}
	} else if token.TokenType == css.StringToken {
		token.Data = bytes.Replace(token.Data, []byte("\\\r\n"), []byte(""), -1)
		token.Data = bytes.Replace(token.Data, []byte("\\\r"), []byte(""), -1)
		token.Data = bytes.Replace(token.Data, []byte("\\\n"), []byte(""), -1)
	}
	return token
}

////////////////////////////////////////////////////////////////

func writeNodes(w io.Writer, nodes []css.Node) error {
	semicolonQueued := false
	for _, n := range nodes {
		if _, ok := n.(*css.NodeToken); semicolonQueued && !ok { // it is only TokenNode for CDO and CDC (<!-- and --> respectively)
			if _, err := w.Write([]byte(";")); err != nil {
				return err
			}
			semicolonQueued = false
		}

		switch m := n.(type) {
		case *css.NodeDeclaration:
			if err := writeDecl(w, m); err != nil {
				return err
			}
			semicolonQueued = true
		case *css.NodeRuleset:
			for i, selGroup := range m.SelGroups {
				if i > 0 {
					if _, err := w.Write([]byte(",")); err != nil {
						return err
					}
				}
				skipSpace := false
				for j, sel := range selGroup.Selectors {
					if len(sel.Nodes) == 1 {
						if token, ok := sel.Nodes[0].(*css.NodeToken); ok {
							if token.TokenType == css.DelimToken && (token.Data[0] == '>' || token.Data[0] == '+' || token.Data[0] == '~') || token.TokenType == css.IncludeMatchToken || token.TokenType == css.DashMatchToken ||
								token.TokenType == css.PrefixMatchToken || token.TokenType == css.SuffixMatchToken || token.TokenType == css.SubstringMatchToken {
								if err := token.Serialize(w); err != nil {
									return err
								}
								skipSpace = true
								continue
							}
						}
					}
					if j > 0 && !skipSpace {
						if _, err := w.Write([]byte(" ")); err != nil {
							return err
						}
					}
					for _, node := range sel.Nodes {
						if err := node.Serialize(w); err != nil {
							return err
						}
					}
					skipSpace = false
				}
			}
			if _, err := w.Write([]byte("{")); err != nil {
				return err
			}
			for i, decl := range m.Decls {
				if i > 0 {
					if _, err := w.Write([]byte(";")); err != nil {
						return err
					}
				}
				if err := writeDecl(w, decl); err != nil {
					return err
				}
			}
			if _, err := w.Write([]byte("}")); err != nil {
				return err
			}
		case *css.NodeAtRule:
			if len(m.Nodes) == 0 && m.Block == nil {
				break
			}
			if err := m.At.Serialize(w); err != nil {
				return err
			}
			skipSpace := false
			for _, token := range m.Nodes {
				if token.TokenType == css.RightParenthesisToken || token.TokenType == css.ColonToken || token.TokenType == css.CommaToken {
					skipSpace = true
				} else if !skipSpace {
					if _, err := w.Write([]byte(" ")); err != nil {
						return err
					}
				} else {
					skipSpace = false
				}
				if token.TokenType == css.LeftParenthesisToken {
					skipSpace = true
				}
				if err := token.Serialize(w); err != nil {
					return err
				}
			}
			if m.Block != nil {
				if err := m.Block.Open.Serialize(w); err != nil {
					return err
				}
				if err := writeNodes(w, m.Block.Nodes); err != nil {
					return err
				}
				if err := m.Block.Close.Serialize(w); err != nil {
					return err
				}
			} else {
				semicolonQueued = true
			}
		case *css.NodeBlock:
			if err := m.Open.Serialize(w); err != nil {
				return err
			}
			if err := writeNodes(w, m.Nodes); err != nil {
				return err
			}
			if err := m.Close.Serialize(w); err != nil {
				return err
			}
		default:
			if err := n.Serialize(w); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeDecl(w io.Writer, decl *css.NodeDeclaration) error {
	if err := decl.Prop.Serialize(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte(":")); err != nil {
		return err
	}
	prevDelim := false
	for j, val := range decl.Vals {
		token, ok := val.(*css.NodeToken)
		currDelim := (ok && (token.TokenType == css.DelimToken || token.TokenType == css.CommaToken || token.TokenType == css.ColonToken))
		if j > 0 && !currDelim && !prevDelim {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}
		if err := val.Serialize(w); err != nil {
			return err
		}
		prevDelim = currDelim
	}
	return nil
}
