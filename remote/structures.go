package remote

// Remote connects to a nuggetFS server and exchanges information.
// Other related packages:
// - remoteFS exposes remote endpoints as a FUSE component.
// - remoteCache introduces a caching layer.

// DataSource represents entities who can be queried about filesystem objects.
type DataSource interface {
}

// DataSink represents entities who can accept data writes.
type DataSink interface {
}
