package common

// Version is the proof-sdk-go module version. Bump on release; surfaced in
// the default User-Agent header so server-side telemetry can attribute traffic
// and emit deprecation notices for old SDK versions.
const Version = "0.3.0"

// UserAgent is the default User-Agent header value sent by AuthenticatedDoer
// when the caller has not already set one on the request.
const UserAgent = "proof-sdk-go/" + Version
