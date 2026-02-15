# Formatted Text Wrapping Test

## Bold Text Across Lines

**Location:** This is a longer location name that should wrap properly without breaking the word Location or any other word in the middle of the text.

**Description:** This demonstrates that bold formatted text at the start of a line correctly tracks column position and wraps subsequent text at word boundaries.

## Inline Formatting

The **quick brown** fox jumps over the *lazy dog* and continues running through a very long sentence that should wrap at word boundaries while preserving all formatting.

## Mixed Formatting in Lists

- **Item Name:** This is a long description that should wrap properly maintaining the indentation and not breaking words in the middle
- **Another Item:** Short description
- **Long Product Name:** This product has a very detailed description that spans multiple lines and should maintain proper alignment throughout the wrapped text

## Task Lists with Formatting

- [ ] **Task One:** Complete the implementation of word wrapping with proper column tracking across formatted text elements
- [x] **Task Two:** Test with various formatting combinations
- [ ] **Urgent:** This is a high priority task with a very long description that needs to wrap properly without breaking any words

## Numbered Lists with Bold

1. **First Step:** Execute the preliminary setup phase which involves several sub-steps and requires careful attention to detail
2. **Second Step:** Proceed with the main implementation
3. **Final Step:** Complete testing and verification of all functionality

## Edge Cases

**VeryLongSingleWordThatExceedsFortyCharacters:** Followed by normal text.

**Normal:** This text should wrap correctly **even with** multiple **bold sections** throughout the paragraph that should all maintain proper column tracking.

## Real-World Example

**Name:** John Smith
**Email:** john.smith@verylongdomainname.example.com
**Address:** 1234 Very Long Street Name Avenue, Apartment Number 567, Building C
**Phone:** +1-555-123-4567

**Order Details:** Customer has ordered multiple items from the catalog including premium products that require special handling and expedited shipping to the specified address.

---

This test file verifies that formatted text (bold, italic) properly tracks column position and wraps at word boundaries without breaking words in the middle.
