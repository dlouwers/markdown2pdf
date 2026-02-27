# Multi-line List Items Test

## Unordered list with wrapped text

- This is a short item
- This is a much longer item that should wrap to multiple lines when rendered in the PDF. The continuation lines should align with the start of the text on the first line, not flush left.
- Another short item
- Yet another very long item with lots of text that will definitely span multiple lines in the rendered PDF output. We need to verify that all continuation lines are properly indented to align with the first line of text.

## Ordered list with wrapped text

1. First item is short
2. Second item is intentionally very long so that it wraps across multiple lines in the PDF output. The continuation lines should be indented to align with the text start position, maintaining proper visual hierarchy.
3. Third item
4. Fourth item with even more text to ensure we test the wrapping behavior thoroughly. This should demonstrate whether the continuation lines maintain proper alignment or if they incorrectly align flush with the left margin.

## Nested lists with wrapping

- Parent item with some longer text that might wrap to demonstrate the alignment in a top-level list item
  - Child item that is also quite long and should wrap to multiple lines while maintaining proper indentation relative to its own bullet point, not the parent
  - Short child
    - Grandchild with very long text to test three levels of nesting. The continuation lines here should align with the grandchild text start, maintaining proper visual hierarchy through all nesting levels.
- Another parent item

## Task list with wrapping

- [x] Completed task with a very long description that spans multiple lines to test whether continuation lines properly align with the first line of text or if they incorrectly flush to the left margin
- [ ] Incomplete task also with lengthy text that will wrap across multiple lines in the rendered PDF, allowing us to verify the alignment behavior for unchecked items
- [x] Another completed item
