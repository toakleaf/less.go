/**
 * Sample LESS Plugin for less.go
 *
 * This plugin demonstrates the structure of LESS plugins.
 * While the plugin infrastructure is in place, custom function
 * execution is still being implemented in less.go.
 *
 * Usage in LESS:
 *   @plugin "sample-plugin.js";
 *   .example { value: my-function(); }
 *
 * The `functions` and `less` objects are provided by the LESS runtime.
 */

// ============================================
// Example Plugin Functions
// ============================================

/**
 * Returns the value of PI
 * Note: LESS has a built-in pi() function, so this would shadow it
 */
functions.add("pi", function () {
  return less.dimension(Math.PI);
});

/**
 * Returns the golden ratio (phi ≈ 1.618)
 */
functions.add("golden-ratio", function () {
  return less.dimension(1.6180339887498949);
});

/**
 * Returns Euler's number (e ≈ 2.718)
 */
functions.add("e-number", function () {
  return less.dimension(Math.E);
});

/**
 * Doubles a numeric value
 * Usage: width: double(10px);
 * Expected: width: 20px;
 */
functions.add("double", function (n) {
  return less.dimension(n.value * 2, n.unit);
});

/**
 * Adds two numbers together
 * Usage: total: add-nums(10, 5);
 * Expected: total: 15;
 */
functions.add("add-nums", function (a, b) {
  return less.dimension(a.value + b.value);
});

/**
 * Returns a brand color
 * Usage: background: brand-color();
 * Expected: background: #4a90d9;
 */
functions.add("brand-color", function () {
  return less.color([74, 144, 217]);
});

/**
 * Creates a color from RGB values
 * Usage: color: make-rgb(255, 128, 0);
 * Expected: color: #ff8000;
 */
functions.add("make-rgb", function (r, g, b) {
  return less.color([r.value, g.value, b.value]);
});

/**
 * Returns a quoted string
 * Usage: content: hello-world();
 * Expected: content: "Hello, World!";
 */
functions.add("hello-world", function () {
  return less.quoted('"', "Hello, World!", true);
});

/**
 * Returns a keyword
 * Usage: display: block-keyword();
 * Expected: display: block;
 */
functions.add("block-keyword", function () {
  return less.keyword("block");
});

/**
 * A custom test function with unique name
 */
functions.add("my-plugin-function", function () {
  return less.dimension(42);
});
