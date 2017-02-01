Nugget is my attempt at a simple network filesystem. This repository stores internal packages and various binaries.

# architecture

## Data storage - Paths/EntryIDs, EntryIDs/Metadata, ChunkIDs/Chunks

All information in the filesystem is stored in 3 key-value stores - stored in 3 files by boltDB.

The first key value store stores a mapping between the file path and a unique ID representing the files metadata. This unique ID is called the EntryID.

The second key value store holds a mapping between EntryID's and Metadata. The metadata stores, among other things, the size of the file, its name, and the ID's of the
chunks that hold its data.

The last key value store holds a mapping between ChunkID's and the data which makes up a chunk. The actual file data is stored here.

### Example operation: read

The filesystem issues a open() followed by a read().

1. Nugget looks up the file's path to get it's entryID. If no such entry exists, the flow aborts with a no-entity error.
2. Nugget looks up the entryID to get the file's metadata, getting out the relevant ChunkIDs so it can access the data.
3. Nugget uses the chunkID's stored in the metadata to request the chunks data, and returns the relevant data back to FUSE.



## Security

Security over the network is achieved by using TLS with both client and server certificate verification. Both client and serve need to be given three files:

1. A PEM-encoded CA certificate to use as the root of trust.
2. A PEM-encoded certificate to be presented to the remote end. This must be signed by the CA certificate.
3. The key for the PEM-encoded certificate.

Both ends check that the remote end presents a certificate which is signed by their root of trust.

Strong (2016) ciphers are used.

# binaries

## nugglocal

`nugglocal` makes a nugget filesystem available locally, without presenting it on the network.

`sudo ./nugglocal /my-mountpoint ~/path-to-dir-where-the-backing-data-should-be-stored/`

## nuggserv

`nuggserv` makes a nugget filesystem available over the network, backing the data in a directory in the local filesystem.

`sudo ./nuggserv --cert nugget.certPEM --cacert ca.pem --key nugget.keyPEM ~/path-to-dir-where-the-backing-data-should-be-stored/`

Note the use of certificates to authenticate to/for clients.

## nugg

`nugg` mounts a nugget network filesystem locally at `mountpoint` using FUSE.

`sudo ./nugg --cert nugget.certPEM --cacert ca.pem --key nugget.keyPEM /mountpoint`

Note the use of certificates to authenticate the server and itself.

# TODO

 - [x] Implement packet encoding for ReadData, Store, Mkdir, Delete
 - [x] Start using FUSE Read() instead of ReadAll() - Make network method to read()
 - [x] Likewise use FUSE Write() and pass it through
 - [ ] Proper tests for nuggdb (Provider)
 - [ ] Prevent remove() from deleting non-empty directories
 - [x] Make nuggdb store chunks as actual files
 - [ ] Make file-backed chunk db delete unneeded folders.
 - [ ] Encode permission information in metadata
 - [ ] Implement ReadData method on the network client
 - [ ] Write tests / documentation for ./packet
 - [ ] Make script to allow incremental backup to S3
 - [ ] Make replication / distributed mode
