---
title: Error Placeholder Test
author: Test Suite
date: 2026-02-28
---

# Error Placeholder Visual Tests

This document tests error placeholder rendering with various error message lengths.

## Test 1: Short Error Message

![Short Error](http://invalid-domain-that-does-not-exist.test/image.png)

## Test 2: Medium Error Message

![Medium length error message that should wrap properly within the placeholder box](http://another-invalid-domain.test/very/long/path/to/image.png)

## Test 3: Long Error Message

![This is a very long alt text that simulates what happens when we have a lengthy error message that needs to wrap multiple times within the error placeholder box to ensure proper rendering](http://extremely-long-domain-name-that-does-not-exist.test/api/v1/images/some-resource-id/thumbnail.jpg)

## Test 4: 404 Error (Real Domain)

![Image not found](https://httpbin.org/status/404)

## Test 5: Network Timeout Simulation

![Timeout test](http://192.0.2.1/image.png)

## Test 6: Multiple Errors in Sequence

![Error A](http://invalid1.test/a.png)
![Error B with longer text that wraps](http://invalid2.test/b.png)
![Error C](http://invalid3.test/c.png)

## Test 7: Error Between Valid Content

Here's some valid text before the error.

![This error is embedded in normal content flow](http://invalid-embedded.test/image.png)

And here's some valid text after the error. This tests that the placeholder doesn't disrupt normal document flow.

## Test 8: Very Long URL

![Short alt](http://this-is-an-extremely-long-domain-name-that-goes-on-and-on.test/api/v2/resources/images/thumbnails/high-resolution/some-uuid-12345678-1234-5678-1234-567812345678.png?quality=high&format=png&size=large)
