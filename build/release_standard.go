// +build !testing,!dev

package build

// Release indicates the kind of release that is built,
// defining timings, amount of (extra) runtime checks.
// Possibilities: standard, testing and dev.
const Release = "standard"
