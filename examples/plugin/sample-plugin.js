/**
 * Sample LESS Plugin for less.go
 *
 * This plugin demonstrates how to create custom functions for LESS.
 * Functions registered here become available in your LESS files when
 * loaded via the @plugin directive.
 *
 * Usage in LESS:
 *   @plugin "sample-plugin.js";
 *   .example { width: double(10px); }
 *
 * The `functions` and `less` objects are provided by the LESS runtime.
 */

// ============================================
// Math Functions
// ============================================

/**
 * Doubles a numeric value
 * Usage: width: double(10px);
 * Output: width: 20px;
 */
functions.add("double", function (n) {
  return less.dimension(n.value * 2, n.unit);
});

/**
 * Adds two numbers together
 * Usage: total: add(10, 5);
 * Output: total: 15;
 */
functions.add("add", function (a, b) {
  return less.dimension(a.value + b.value);
});

/**
 * Returns the square root of a number
 * Usage: value: sqrt-val(16);
 * Output: value: 4;
 */
functions.add("sqrt-val", function (n) {
  return less.dimension(Math.sqrt(n.value), n.unit);
});

/**
 * Multiplies a value by a factor
 * Usage: width: multiply(10px, 3);
 * Output: width: 30px;
 */
functions.add("multiply", function (n, factor) {
  return less.dimension(n.value * factor.value, n.unit);
});

/**
 * Clamps a value between min and max
 * Usage: width: clamp-val(50px, 100px, 200px);
 * Output: width: 100px;
 */
functions.add("clamp-val", function (min, val, max) {
  var result = Math.min(Math.max(min.value, val.value), max.value);
  return less.dimension(result, val.unit || min.unit);
});

// ============================================
// Color Functions
// ============================================

/**
 * Returns the brand color
 * Usage: background: brand-color();
 * Output: background: #4a90d9;
 */
functions.add("brand-color", function () {
  return less.color([74, 144, 217]);
});

/**
 * Creates a color from RGB values
 * Usage: color: make-rgb(255, 128, 0);
 * Output: color: #ff8000;
 */
functions.add("make-rgb", function (r, g, b) {
  return less.color([r.value, g.value, b.value]);
});

/**
 * Returns a lightened version of the brand color
 * Usage: background: lighten-brand(20);
 * Output: background: #a5c8ed; (approximately)
 */
functions.add("lighten-brand", function (amount) {
  var percent = amount.value / 100;
  var lighten = function (c) {
    return Math.min(255, Math.round(c + (255 - c) * percent));
  };
  return less.color([lighten(74), lighten(144), lighten(217)]);
});

// ============================================
// String Functions
// ============================================

/**
 * Returns a greeting string
 * Usage: content: greet("World");
 * Output: content: "Hello, World!";
 */
functions.add("greet", function (name) {
  return less.quoted('"', "Hello, " + name.value + "!", true);
});

/**
 * Adds a vendor prefix to a property name
 * Usage: content: prefix("transform");
 * Output: content: "-webkit-transform";
 */
functions.add("prefix", function (prop) {
  return less.quoted('"', "-webkit-" + prop.value, true);
});

/**
 * Joins multiple values with a separator
 * Usage: content: join-str("-", "a", "b", "c");
 * Output: content: "a-b-c";
 */
functions.add("join-str", function (sep) {
  var parts = [];
  for (var i = 1; i < arguments.length; i++) {
    parts.push(arguments[i].value);
  }
  return less.quoted('"', parts.join(sep.value), true);
});

// ============================================
// Keyword Functions
// ============================================

/**
 * Returns a keyword value
 * Usage: display: block-keyword();
 * Output: display: block;
 */
functions.add("block-keyword", function () {
  return less.keyword("block");
});

/**
 * Returns true or false as a keyword
 * Usage: @is-large: is-greater(100, 50);
 * Output: @is-large: true;
 */
functions.add("is-greater", function (a, b) {
  return less.keyword(a.value > b.value ? "true" : "false");
});

// ============================================
// Stateful Functions (Collection Example)
// ============================================

var collection = [];

/**
 * Stores a value in the collection
 * Usage: store(42); store("hello");
 * Note: Returns false (no output)
 */
functions.add("store", function (val) {
  collection.push(val);
  return false; // No output
});

/**
 * Returns all stored values
 * Usage: values: list();
 * Output: values: 42, "hello", ...;
 */
functions.add("list", function () {
  return less.value(collection);
});

/**
 * Clears the collection
 * Usage: clear();
 */
functions.add("clear", function () {
  collection = [];
  return false;
});

// ============================================
// Utility Functions
// ============================================

/**
 * Returns the current timestamp
 * Usage: cache-bust: timestamp();
 * Output: cache-bust: 1234567890123;
 */
functions.add("timestamp", function () {
  return less.dimension(Date.now());
});

/**
 * Returns the current year
 * Usage: copyright-year: year();
 * Output: copyright-year: 2025;
 */
functions.add("year", function () {
  return less.dimension(new Date().getFullYear());
});
