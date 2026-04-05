# src2tex-go <small>v2.236</small>

プログラムのソースコードを XeLaTeX / LuaLaTeX / upLaTeX / pdfLaTeX 互換のLaTeX 文書に変換する CLI ツールです。コメントにTeXの数式を書くことができ、動くソースコードをそのまま文書化できます。例えば、サンプルコードに解説を含めて、すぐに学生向けのプリント教材を作成できます。

1992年に城西大学理学部に所属していた天野一男先生（数学）が開発した C プログラム `src2tex` の version 2.12 をベースに Go言語で完全に書き直したバージョンです。

バージョン番号は √5 の小数表現で、バージョンアップの都度、近似していく形をとっています。現在のバージョンは、 **version 2.236** です。

前回のバージョンはオリジナル版に近い挙動を目指した移植を目的としていましたが、今回は現在のLaTeX事情を加味し、モダンLaTeXの使用を前提とするアーキテクチャの変更を行いました。詳しくは ARCHITECTURE.md および NOTE.md に記載しました。

## 前提環境

| ツール | 用途 |
|--------|------|
| [XeLaTeX](https://tug.org/xetex/) など$^1$（TeX Live / MacTeX） | PDF 生成 |
| [Ghostscript](https://www.ghostscript.com/) (`gs`) | EPS → PDF 変換 |

<small>1: ほかにも LuaLaTeX、upLaTeX、pdfTeX（英語文書のみ）をサポートしています。プレーンなTeXのサポートは前回の [互換バージョン](https://github.com/yupyom/src2tex-go) で実現していましたが、著者自身がプレーンのTeXは使わないので、今回はLaTeX系のみにしました。</small>

## インストール

Go 1.21 以上が必要です。

```bash
go install github.com/yupyom/src2tex-go-v2.236/cmd/src2tex@latest
```

またはリポジトリをクローンしてビルド:

```bash
git clone <repo-url>
cd src2tex-go-v2.236
go build -o src2tex ./cmd/src2tex/
```

## 基本的な使い方

```bash
# Go ファイルを変換（hanoi.go.tex が生成される）
src2tex hanoi.go

# 言語を明示指定
src2tex -lang reduce popgen.red

# stdin → stdout
cat hanoi.go | src2tex -lang go

# PDF 生成（testdata/input/ から実行）
cd testdata/input
xelatex -halt-on-error -interaction=nonstopmode hanoi.go.tex
```

なお、出力ファイル名は `<入力ファイル名.拡張子>.tex`のように、拡張子を保持して `.tex` を付加したものになります。

### 対応言語一覧

| 種別 | 言語 | 拡張子 | `-lang` 値 | キーワード太字 | docstring |
|------|------|--------|------------|:-:|:-:|
| C 系 | C | `.c`, `.h` | `c` | ✅ | — |
| | Go | `.go` | `go` | ✅ | — |
| | Java | `.java` | `java` | ✅ | — |
| | C++ | `.cpp`, `.cc`, `.cxx`, `.hpp` | `cpp` | ✅ | — |
| | C# | `.cs` | `csharp` | ✅ | — |
| | Dart | `.dart` | `dart` | ✅ | — |
| | JavaScript | `.js`, `.mjs` | `js` | ✅ | — |
| | TypeScript | `.ts`, `.tsx` | `ts` | ✅ | — |
| | Rust | `.rs` | `rust` | ✅ | — |
| | Kotlin | `.kt`, `.kts` | `kotlin` | ✅ | — |
| | Swift | `.swift` | `swift` | ✅ | — |
| Hash 系 | Shell | `.sh`, `.bash` | `sh` | ✅ | — |
| | Python | `.py` | `python` | ✅ | ✅ (`"""` / `'''`) |
| | Ruby | `.rb` | `ruby` | ✅ | — |
| | Perl | `.pl`, `.pm` | `perl` | ✅ | — |
| | Makefile $^2$ | `.mk`, `Makefile`* | `make` | — | — |
| | Tcl | `.tcl` | `tcl` | ✅ | — |
| Percent 系 | REDUCE | `.red` | `reduce` | ✅ | — |
| | MATLAB/Octave | `.m` | `matlab` | — | — |
| Semicolon 系 | Lisp/Scheme | `.lisp`, `.scm`, `.el` | `lisp` | ✅ | — |
| Pascal 系 | Pascal | `.pas`, `.p` | `pascal` | ✅ | — |
| XML 系 | XML | `.xml`, `.xsl`, `.xslt`, `.svg`, `.xhtml` | `xml` | — | — |
| CSS 系 | CSS | `.css` | `css` | — | — |
| HTML 系 | HTML | `.html`, `.htm`, `.php` | `html` | ✅ (タグ名) | — |

<small>2: `Makefile`, `makefile`, `GNUmakefile` は拡張子なしでもファイル名で自動判別されます。</small>


### コマンドラインオプション

| オプション | 説明 |
|-----------|------|
| `-lang <name>` | 言語を明示指定（拡張子で判別できない場合） |
| `-o <file>` | 出力ファイルパス（デフォルト: `<入力>.tex`） |
| `-font <name>` | コード部の等幅フォント（`\setmonofont` に設定、デフォルト: `CMU Typewriter Text`、TeX Live 未検出時は `Courier New`） |
| `-commentfont <name>` | コメント部の CJK 明朝フォント（デフォルト: 自動検出） |
| `-fontdir <path>` | フォントインストール先（デフォルト: `~/.src2tex/fonts/`） |
| `-listfonts` | 利用可能なフォント一覧を表示して終了 |
| `-header <file>` | `\fancyhead[R]{...}` をファイル内容で置換 |
| `-footer <file>` | `\fancyfoot[R]{...}` をファイル内容で置換 |
| `-linenumbers` | ソース行番号を左マージンに表示 |
| `-tab <n>` | タブ幅をスペース数で指定（デフォルト: 8） |
| `-engine <name>` | TeX エンジン（`xelatex`, `lualatex`, `uplatex`, `pdflatex`, `tectonic`、デフォルト: `xelatex`） |
| `-paper <size>` | 用紙サイズ（`a4`, `b5`, `letter`、デフォルト: `a4`） |


## LaTeX エンジンごとのテンプレートについて

`src2tex engine init` を実行すると、プリアンブルテンプレートをユーザーディレクトリ `~/.src2tex/engines/` に展開します。プロジェクトに合わせてカスタマイズできます。

### 対応エンジン

| エンジン | `-engine` 値 | CJK 対応 | fontspec | 特徴 |
|---------|-------------|:--------:|:--------:|------|
| XeLaTeX | `xelatex` (デフォルト) | ✅ | ✅ | fontspec + xeCJK |
| Tectonic | `tectonic` | ✅ | ✅ | XeTeX 互換、fontspec対応 |
| LuaLaTeX | `lualatex` | ✅ | ✅ | luatexja-fontspec |
| upLaTeX | `uplatex` | ✅ | ❌ | jsarticle + otf、dvipdfmx パイプライン |
| pdfLaTeX | `pdflatex` | ✅ | ❌ | bxjsarticle + bxcjkjatype（`ja=standard`） |

### 使用例

```bash
# テンプレートを ~/.src2tex/engines/ に展開
src2tex engine init

# 利用可能なエンジン一覧
src2tex engine list

# pdfLaTeX で欧文のみのソースを変換
src2tex -engine pdflatex -paper letter hello.c

# LuaLaTeX で日本語ソースを変換
src2tex -engine lualatex report.go

# upLaTeX + dvipdfmx（.dvi → .pdf の手動変換が必要）
src2tex -engine uplatex report.go
uplatex report.go.tex && dvipdfmx report.go.dvi

# B5 用紙で変換
src2tex -paper b5 report.go
```

テンプレートファイルは `~/.src2tex/engines/<engine>/preamble.tmpl` にあり、
Go の `text/template` 構文（デリミタ: `<% %>`）で記述されています。

### テンプレートの保護（`auto` フラグ）

`engine init` 時、各エンジンの `engine.json` に `"auto": true` が自動設定されます。この `auto` フラグによって、ユーザーの編集内容が `engine init` で上書きされるのを防止できます。

| `engine.json` の `auto` | `engine init` 時の動作 |
|---|---|
| `true` | ビルトインテンプレートで**上書き**される（最新化） |
| `false` または省略 | **スキップ**される（ユーザー編集を保持） |

ビルトインエンジンのテンプレートをカスタマイズしたい場合は、対象の `engine.json` から `"auto": true` を削除（または `false` に変更）してから、`preamble.tmpl` を編集してください。

### カスタムテンプレートの作成

既存エンジン設定をコピーして独自のエンジンを作成できます。別のTeXエンジンを使いたいときや、テンプレート自体をカスタマイズしたいときにご利用ください。なお、ご自身で作る場合は、上述のように `engine.json` の `auto` フラグを `false` にするようにしてください。

```bash
# LuaLaTeX をベースにカスタムテンプレートを作成
cp -r ~/.src2tex/engines/lualatex ~/.src2tex/engines/my-lualatex

# engine.json を編集（auto を消して名前を変更）
vi ~/.src2tex/engines/my-lualatex/engine.json

# preamble.tmpl を自由にカスタマイズ
vi ~/.src2tex/engines/my-lualatex/preamble.tmpl

# カスタムエンジンで変換
src2tex -engine my-lualatex report.go
```

カスタムエンジンは `engine list` で `[custom]` ラベル付きで表示されます:
```
Available engines:
  lualatex     Unicode TeX engine with luatexja (CJK support)
  xelatex      Unicode TeX engine with fontspec + xeCJK (default)
  my-lualatex  My customized LuaLaTeX [custom]
```

> **注意**: ユーザー作成のカスタムエンジンには `"auto": true` を**付けないでください**。`auto` を付けると、`engine init` でビルトインテンプレートに上書きされる可能性があります。

### エンジン別検証方法

お使いの環境で変換したLaTeXファイルのコンパイルが通るかどうかは、以下で確認ができます。事前に [Task](https://taskfile.dev/) を導入しておいてください。

```bash
# 全エンジンの検証を実行
task verify:all

# 個別エンジンの検証
task verify:xelatex
task verify:pdflatex
task verify:lualatex
task verify:uplatex
task verify:tectonic
```

## フォント管理について

fontspec対応のエンジンでは、コード部・コメント部のフォントをカスタマイズできます。推奨書体は、全角・半角の等幅フォントで、特にスペースベースのソースコードの整形に適しています。著者が推奨する書体を簡単にインストールできるユーティリティコマンドを用意していますが、ご自身でインストールした書体を使うこともできます。

```bash
# 利用可能なフォント一覧
src2tex font list

# フォントをインストール（GitHub からダウンロード）
src2tex font install hackgen
src2tex font install all

# コメント用フォントをインストール
src2tex font install-comment haranoaji
src2tex font install-comment all

# インストール済みフォントで変換
src2tex -font hackgen -commentfont haranoaji hanoi.go
```

フォントは `~/.src2tex/fonts/` にインストールされます。`-fontdir` で変更可能です。

### フォント設定ファイル（`~/.src2tex/fonts.json`）

フォントのインストール時、または `src2tex font init` コマンド実行時に、`~/.src2tex/fonts.json` が自動生成されます。このファイルは、利用中の全書体の一覧とそのパス情報を記載したリファレンス兼設定ファイルです。

```bash
# fonts.json を生成/更新
src2tex font init
```

#### ファイル構造

```json
{
  "_comment": "src2tex font configuration. Entries with \"auto\": true are system-managed and regenerated on init. Remove \"auto\" or set it to false to prevent overwriting.",
  "code_fonts": [
    {
      "name": "hackgen",
      "display_name": "HackGen",
      "license": "SIL OFL",
      "unified": true,
      "regular_file": "HackGen-Regular.ttf",
      "bold_file": "HackGen-Bold.ttf",
      "font_dir": "/Users/xxx/.src2tex/fonts/hackgen",
      "installed": true,
      "auto": true,
      "description": "Hack + 源柔ゴシック。プログラミング向けの人気フォント。"
    }
  ],
  "comment_fonts": [
    {
      "name": "haranoaji",
      "display_name": "原ノ味明朝",
      "license": "SIL OFL",
      "regular_file": "HaranoAjiMincho-Regular.otf",
      "bold_file": "HaranoAjiMincho-Bold.otf",
      "font_dir": "/usr/local/texlive/2026/texmf-dist/fonts/opentype/public/haranoaji",
      "texlive": true,
      "installed": true,
      "auto": true,
      "description": "TeX Live 同梱の高品質明朝体。ダウンロード不要。"
    }
  ]
}
```

#### 各フィールドの説明

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `name` | string | **必須**。`-font` / `-commentfont` オプションで指定する名前 |
| `display_name` | string | 表示用の名前。`font list` で使用 |
| `license` | string | ライセンス種別（参考情報） |
| `unified` | bool | `true` の場合、欧文と CJK が統合されたフォント。`\setCJKmonofont` も同じフォントで設定される |
| `regular_file` | string | **必須**。レギュラーウェイトのフォントファイル名 |
| `bold_file` | string | ボールドウェイトのフォントファイル名。省略時はボールド指定なし |
| `font_dir` | string | フォントファイルが格納されているディレクトリの絶対パス。省略時は `~/.src2tex/fonts/{name}/` |
| `texlive` | bool | `true` の場合、TeX Live に同梱されているフォント（ダウンロード不要） |
| `installed` | bool | 自動検出されたインストール状態。手動編集する必要はない |
| `auto` | bool | **重要**。`true` の場合、システム管理エントリ。`font init` 時にビルトイン定義から再生成される。`false` または省略時はユーザー編集エントリとして再生成で上書きされない |
| `description` | string | 短い説明文（参考情報） |

#### `auto` フラグによる保護の仕組み

`src2tex font init` や `src2tex font install` 実行時の再生成ポリシー:

| `auto` の値 | `font init` 時の動作 |
|---|---|
| `true` | ビルトイン定義から再生成される（`font_dir`, `installed` 等が最新化） |
| `false` または省略 | **上書きされない**。名前がビルトインと同じでも保持される |

ビルトイン書体の設定をカスタマイズしたい場合は、対象エントリの `"auto": true` を削除（または `false` に変更）してから、`font_dir` などを編集してください。

#### パス解決ルール

`font_dir` と `regular_file` の組み合わせでフォントの実体パスが決定されます:

1. **`font_dir` が指定されている場合** → `{font_dir}/{regular_file}` がフォントパスになる
2. **`font_dir` が省略されている場合** → `~/.src2tex/fonts/{name}/{regular_file}` が使用される
3. **`regular_file` にパス区切り（`/`）が含まれる場合** → その `dirname` 部分がフォントディレクトリとして使用される（`font_dir` より優先）

生成される LaTeX コマンド例:
```latex
\setmonofont[Path=/Users/xxx/.src2tex/fonts/hackgen/, BoldFont=HackGen-Bold.ttf]{HackGen-Regular.ttf}
```

#### カスタムフォントの追加例

OS にインストールされたフォントや、任意のディレクトリのフォントを使うことができます。`code_fonts` 配列に新しいエントリを追加してください:

```json
{
  "name": "myricam",
  "display_name": "MyricaM",
  "license": "SIL OFL",
  "unified": true,
  "regular_file": "MyricaM-Regular.ttf",
  "bold_file": "MyricaM-Bold.ttf",
  "font_dir": "/Library/Fonts",
  "installed": true
}
```

追加後は `-font myricam` で指定できます。`src2tex font init` を再実行しても、`"auto"` がない（または `false` の）エントリは保持されます。

> **注意**: ユーザーが手動で追加するエントリには `"auto": true` を**付けないでください**。`auto` を付けると、次回の `font init` でビルトイン定義で上書きされる可能性があります。

> **注意**: `fontspec` を使用するエンジン（XeLaTeX, LuaLaTeX, Tectonic）でのみフォント指定が有効です。upLaTeX / pdfLaTeX では `-font` オプションは無視されます。

### ダウンロード時のプロキシ設定

企業ネットワーク等でプロキシが必要な場合、環境変数で設定してください:

```bash
# Linux / macOS
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
```

```powershell
# Windows (PowerShell)
$env:HTTP_PROXY = "http://proxy.example.com:8080"
$env:HTTPS_PROXY = "http://proxy.example.com:8080"
```

```bat
:: Windows (cmd)
set HTTP_PROXY=http://proxy.example.com:8080
set HTTPS_PROXY=http://proxy.example.com:8080
```


## TeX コメント記法について

コメント内で `{\` から始まるブロックは、LaTeX コマンドとして直接出力されます。

### インライン記法

コメントの末尾に TeX コマンドを埋め込む:

```c
/* 半径 $r = \sqrt{x^2 + y^2}$ */
```

```python
# 積分 $\int_0^1 f(x)\,dx$
```

```pascal
{ Euclid のアルゴリズム {\ \hfill --- 最大公約数} }
```

### `{\  }` 記法（インライン TeX ブロック）

コメント行の中に `{\ ... }` を書くと、そのブロックが生の LaTeX として出力されます:

```python
# {\ テキスト \hfill}
# {\ \hrulefill}
# {\ \begin{quote} インデントされたテキスト \end{quote} }
```

### `{\null ... }` 記法（複数行 TeX ブロック）

`{\null` で始まるブロックは複数行にまたがる LaTeX 出力に使えます:

```python
# {\null
# \begin{eqnarray}
#   f(x) &=& x^2 + 1 \\
#   g(x) &=& \sqrt{x}
# \end{eqnarray}
# }
```

Pascal では `{ }` がブロックコメント記号を兼ねるため、`{\ ... }` がそのままパススルーとして機能します。

### Python docstring（トリプルクオート）

Python の `"""..."""`（または `'''...'''`）は、行頭に出現した場合にブロックコメントとして扱われます。
内部で `{\ ... }` 記法も使えます:

```python
"""
{\ \centerline{\bf モジュール概要} }
このモジュールはハノイの塔を解きます。
"""

def hanoi(n, src, dst, tmp):
    '''n 枚のディスクを移動する。'''
```

行頭以外の `"""` はコメントとして扱われません（文字列リテラルとして処理されます）。

## カスタムヘッダー / フッターについて

```bash
# ヘッダーを置換
printf '\\fancyhead[L]{\\rm My Project}\n' > myheader.tex
src2tex -header myheader.tex hanoi.go

# フッターを置換
printf '\\fancyfoot[C]{\\thepage}\n' > myfooter.tex
src2tex -footer myfooter.tex hanoi.go
```

ファイルには `\fancyhead` / `\fancyfoot` コマンドをそのまま記述します。
`\usepackage{fancyhdr}` と `\pagestyle{fancy}` はプリアンブルに自動出力されるため不要です。


## ライセンス

MITライセンスとして、現状有姿、無保証で配布します。ご自由にお使いください。

オリジナル版の `src2tex` version 2.12 は、天野一男（Kazuo AMANO）氏・野本慎一（Shinichi NOMOTO）氏による作品で、かつて、フリーソフトウェアとして公開されていました。
