# Word Wrapping Test

## Short Lines

This is a short line that fits easily within the thermal printer width.

## Long Lines That Should Wrap

This is a much longer line that contains significantly more text than can fit on a single line of a thermal printer which typically has a width of approximately forty-six characters per line and should wrap automatically at word boundaries without breaking words in the middle.

## Multiple Paragraphs

The quick brown fox jumps over the lazy dog. This sentence is repeated to demonstrate paragraph wrapping behavior.

The quick brown fox jumps over the lazy dog again. This shows that each paragraph is wrapped independently and maintains proper spacing.

## Edge Cases

Supercalifragilisticexpialidocious is an extremely long word that exceeds the line width.

Short words like: a an the it is at on in or and but for nor yet so.

## Mixed Content

**Bold text wrapping:** This bold text should also wrap properly at word boundaries and maintain the bold formatting throughout the wrapped lines without breaking in the middle of words.

*Italic text wrapping:* Similarly, this italic text (rendered as underline) should wrap correctly and preserve formatting across line breaks.

## List Wrapping

- This is a bullet point with a very long description that should wrap properly while maintaining the indentation and bullet point alignment throughout all wrapped lines.
- Short item
- Another long bullet point that contains enough text to demonstrate wrapping behavior with proper indentation for continuation lines in the list structure.

1. Numbered items also need proper wrapping to ensure that long descriptions maintain alignment with the number prefix throughout all continuation lines.
2. Short numbered item
3. This demonstrates that ordered lists handle word wrapping correctly without breaking words and maintain proper indentation alignment.

## Technical Terms

Electroencephalographically and pneumonoultramicroscopicsilicovolcanoconiosis are examples of very long technical terms.

Common technical terms: implementation, configuration, synchronization, initialization, internationalization.

---

**Note:** This test file helps verify that text wraps at word boundaries (spaces) rather than breaking words in the middle, maintaining readability on thermal printer output.
