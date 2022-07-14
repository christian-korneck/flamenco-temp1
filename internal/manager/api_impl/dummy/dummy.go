// Package dummy contains non-functional implementations of certain interfaces.
// This allows the Flamenco API to be started with a subset of its
// functionality, so that the API can be served without Shaman file storage, or
// without the persistence layer.
//
// This is used for the first startup of Flamenco, where for example the shared
// storage location isn't configured yet, and thus the Shaman shouldn't start.
package dummy
