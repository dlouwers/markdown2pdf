# Remote Images Test

This document tests HTTP/HTTPS image fetching in markdown2pdf.

## PNG from URL

Remote PNG image (Go gopher logo):

![Go Gopher](https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png)

## JPEG from URL

Remote JPEG image (placeholder):

![JPEG Image](https://picsum.photos/400/300.jpg)

## SVG from URL

Remote SVG image (Go gopher):

![SVG Gopher](https://go.dev/images/gophers/ladder.svg)

## Mixed Local and Remote

Local image:

![Local Test](test.png)

Remote image:

![Remote Logo](https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Aqua.png)

## Error Handling

### 404 Not Found

This should render a placeholder:

![Missing Remote](https://example.com/this-does-not-exist-12345.png)

### Invalid URL

This should render a placeholder:

![Invalid](https://invalid-domain-that-does-not-exist-12345.com/image.png)
