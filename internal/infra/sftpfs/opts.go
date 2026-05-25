package sftpfs

const (
	// SFTPMaxPacket is the maximum packet size negotiated with the server.
	// 32 KiB is the SFTPv3 ceiling, larger values get rejected by OpenSSH.
	SFTPMaxPacket = 32 * 1024

	// SFTPMaxConcurrentReqs caps the number of in-flight read/write packets
	// per file. 64 saturates gigabit on LAN; on high-latency WAN the user can
	// raise it via settings later.
	SFTPMaxConcurrentReqs = 64
)
