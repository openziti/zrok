package controller

// rateLimitedRetryAfter is the Retry-After, in seconds, answered when the ziti controller is rate
// limiting. ziti's 429 carries no hint; five seconds is a guess that keeps a client from hammering a
// controller that is already shedding load.
const rateLimitedRetryAfter = 5
