# Configuration File Format

The configuration file is in YAML format with the following structure:

## Basic Structure

```yaml
surround-regexp:
  - start: "start_pattern"
    endings: ["end_pattern_1", "end_pattern_2"]
  - start: "another_pattern"
    end: "single_end_pattern"

form-prefix-regexp:
  - "prefix_pattern"

simple-label-regexp:
  - "label_pattern"

compound-label-regexp:
  - "compound_pattern"

operator-regexp:
  - pattern: "operator_pattern"
    prefix-prec: 100
    infix-prec: 50
    postfix-prec: 75

bracket-regexp:
  - start: "start_pattern"
    endings: ["end_pattern_1", "end_pattern_2"]
    infix: bool
    outfix: bool
```

## Pattern Types

### 1. Surround Patterns (`surround-regexp`)

Surround patterns have three components, namely:

- `start`, which is a single regular expression, which must match the whole
  of a token's text. Required.
- `endings`, which is a list of substitutions, where $0 is replaced by the
  whole of the token's text and $1, $2, etc by any captured group. To generate
  a `$` use `$$`.
- `end`, which is a single regular expression, which must match the whole of a
  token's text. Optional - although one of `endings` and `end` must be present.

The rules for using these components are as follows:

1. If a token matches the `start` expression in full then it is considered a
   form-start. 
2. If `endings` is present, then triggered by a match of a start token, all the
   substitutions are generated and these are the matching form-end for that
   start token.
3. If `end` is present, then any token matching the `end` expression in 
   full is additionally considered to be a form-end token. 
    - When `endings` are missing, then _in addition_ all of these tokens 
      are considered to be the form-end for the corresponding form-start.
    - When `endings` are present, then _only_ the generated tokens are
      considered pairs i.e. `endings` refines the pairing relationship.
4. When `end` is missing, `endings` are required and an attempt is made to
   synthesize the `end` from the endings. This can be done when the 
   substitutions are either constant or only include $0 and not $1, $2, ...
    - If the substitution text includes $N, where N != 1, re-classify
      will fail with an error.



Here is a simplified example, where `def` is matched with `enddef` or `end`;
`if` and `while` are matched with `endif`/`if_end` and `endwhile`/`white_end`
respectively, and `begin` is matched with `end`.
```yaml
surround-regexp:
  - start: "def"
    endings: ["enddef", "end"]
  - start: "if|while"
    endings: ["end$0", "$0_end"]  # $0 substitutes the matched text
  - start: "begin"
    end: "end"  # Single end pattern (alternative to endings array)
```

### 2. Form Prefix Patterns (`form-prefix-regexp`)

This introduces a list of regexs for identifying form-prefixes.

```yaml
form-prefix-regexp:
  - "not"
  - "\\+"  # Literal + character
```

### 3. Simple Label Patterns (`simple-label-regexp`)

This introduces a list of regexs for identifying simple labels:

```yaml
simple-label-regexp:
  - "[a-z]+:"
  - "label[0-9]+"
```

### 4. Compound Label Patterns (`compound-label-regexp`)

This introduces a list of regexs for identifying compound labels:

```yaml
compound-label-regexp:
  - "elseif"
  - "[a-z]+__[a-z]+"
```

### 5. Operator Patterns (`operator-regexp`)

Operators with precedence values for prefix, infix, and postfix positions:

```yaml
operator-regexp:
  - pattern: "="
    prefix-prec: 0
    infix-prec: 1
    postfix-prec: 0
  - pattern: "\\+\\+"
    prefix-prec: 100
    infix-prec: 0
    postfix-prec: 75
```

### 6. Bracket Patterns (`bracket-regexp`)

Bracket patterns define opening delimiters and their matching closing delimiters. The configuration includes flags to specify whether the brackets can be used in infix or outfix contexts:

```yaml
bracket-regexp:
  - start: "\\("
    endings: ["\\)"]
    infix: true
    outfix: true
  - start: "\\["
    endings: ["\\]"]
    infix: true
    outfix: false
  - start: "\\{"
    endings: ["\\}"]
    infix: false
    outfix: true
```

The fields are:
- `start`: Regular expression matching the opening bracket
- `endings`: List of possible closing bracket patterns (does NOT support $0 substitution)
- `infix`: Boolean indicating if the bracket can be used in infix position (e.g., `f[x]`, `f(x)`)
- `outfix`: Boolean indicating if the bracket can be used in outfix position (e.g., `(a, b)`, `{a := b}`)


## Example

In this simple example we pair `if`/`fi` together and `while`/`done` together
and make `do` a simple label.

### Configuration

```yaml
surround-regexp:
  - start: if
    endings: fi
  - start: while
    endings: done

simple-label-regexp:
  - do
```

### Testing

Create a simple test with the existing test configuration (one token per line):

```bash
printf "if\nvariable\nendif\n" | re-classify config.yaml
```

Expected output:
```txt
S endif if_end
V
E
```

This shows:
- `if` classified as Start token (`S`) with expected endings `endif` and `if_end`
- `variable` classified as Variable (`V`)  
- `endif` classified as End token (`E`)
