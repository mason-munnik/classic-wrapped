// Package musicbrainz matches classified composer/performer credits against
// the MusicBrainz API to resolve canonical work/recording IDs.
//
// Not yet implemented on the Python side either — this is the next pipeline
// stage. Good fit for goroutines + a rate limiter: MusicBrainz's
// unauthenticated tier asks for roughly one request per second.
package musicbrainz
