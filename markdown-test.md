<!--toc:start-->
- [Markdown Parser Test Suite](#markdown-parser-test-suite)
  - [Paragraphs](#paragraphs)
  - [Headings](#headings)
- [Heading 1](#heading-1)
  - [Heading 2](#heading-2)
    - [Heading 3](#heading-3)
      - [Heading 4](#heading-4)
        - [Heading 5](#heading-5)
          - [Heading 6](#heading-6)
  - [Blockquotes](#blockquotes)
  - [Lists](#lists)
    - [Unordered](#unordered)
    - [Ordered](#ordered)
    - [Mixed](#mixed)
    - [Task List](#task-list)
  - [Links](#links)
  - [Wikilinks](#wikilinks)
  - [Tags](#tags)
  - [Images](#images)
  - [Code](#code)
    - [Fenced Code Block](#fenced-code-block)
    - [Fenced Code Block Without Language](#fenced-code-block-without-language)
    - [Indented Code Block](#indented-code-block)
    - [Nested Fence in Blockquote](#nested-fence-in-blockquote)
  - [Tables](#tables)
  - [Horizontal Rules](#horizontal-rules)
  - [Footnotes](#footnotes)
  - [Definition List](#definition-list)
  - [HTML Blocks and Inline HTML](#html-blocks-and-inline-html)
  - [Escaping and Entities](#escaping-and-entities)
  - [Special Characters and Unicode](#special-characters-and-unicode)
  - [Nested Structures](#nested-structures)
  - [Reference-Style Shortcuts](#reference-style-shortcuts)
  - [Front Matter](#front-matter)
  - [Math-Like Content](#math-like-content)
  - [Raw Text Edge Cases](#raw-text-edge-cases)
  - [Final Section](#final-section)
<!--toc:end-->

# Markdown Parser Test Suite

This file exercises a wide range of Markdown elements and common extensions.

[TOC]

## Paragraphs

This is a normal paragraph with **bold**, *italic*, ***bold italic***, ~~strikethrough~~, ==highlight==, and `inline code`.

Line with a soft break  
and a hard break.

A second paragraph with escaped characters: \* \_ \# \[ \] \( \) \\.

---

## Headings


# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

---

## Blockquotes

> Simple blockquote.
>
> Multi-paragraph blockquote.
>
> > Nested blockquote.
>
> Back to the first level.

---

## Lists

### Unordered

- Item one
- Item two
  - Nested item two.a
  - Nested item two.b
    - Deeply nested item
- Item three

### Ordered

1. First item
2. Second item
   1. Nested ordered item
   2. Another nested ordered item
3. Third item

### Mixed

1. Ordered item
   - Nested unordered item
   - Another nested unordered item
2. Ordered item two

### Task List

- [x] Completed task
- [ ] Incomplete task
- [ ] Another task with `code`

---

## Links

Inline link: [OpenAI](https://openai.com)

Autolink: <https://example.com>

Email autolink: <user@example.com>

Reference link: [Example][example-ref]

[example-ref]: https://example.org "Example reference title"

---

## Code

Inline code example: `const x = 42;`

This is a test

### Fenced Code Block

```python
def greet(name: str) -> str:
    return f"Hello, {name}!"

print(greet("Markdown"))
```

### Fenced Code Block Without Language

```
plain text code fence
<not parsed as html>
alright
```

### Indented Code Block

    function add(a, b) {
        return a + b;
    }

### Nested Fence in Blockquote

> ```json
> {
>   "name": "example",
>   "enabled": true
> }
> ```

---

## Tables

| Column A | Column B | Column C |
|---------:|:--------:|----------|
| Right    | Center   | Left     |
| 123      | `code`   | **bold** |
| Cell 3   | Cell 4   | Cell 5   |


---
## Horizontal Rules

---

***

___

---

## Footnotes

Here is a sentence with a footnote.[^1]

Here is another footnote reference.[^long-note]

[^1]: This is the first footnote.
[^long-note]: This is a longer footnote
    that continues on a new indented line,
    and can include **formatting** and `code`.

---


## HTML Blocks and Inline HTML

<div class="note">
  <p>This is an HTML block inside Markdown.</p>
</div>

Inline HTML: <span data-test="inline-html">inline span</span>

<!-- This is an HTML comment -->

---

## Escaping and Entities

Escaped punctuation: \*not italic\* and \[not a link](#).

Entities: &copy; &amp; &lt; &gt;

---

## Nested Structures

> Blockquote containing a list:
> - Item A
> - Item B
>   1. Nested ordered
>   2. Nested ordered two
>
> And a table-like line:
> | not | necessarily | parsed |

- List item containing a blockquote:
  > Quoted inside a list.

- List item containing code:

  ```bash
  echo "inside a list"
  ```

---

## Math-Like Content

Inline pseudo-math: $E = mc^2$

Block pseudo-math:

$$
\int_0^1 x^2 \, dx = \frac{1}{3}
$$
