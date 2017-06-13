package data

import (
	"bufio"
	"errors"
	"fmt"
	"me/vilsol/api/utils"
	"net"
	"strconv"
	"time"
)

type FactorioServerDataModel struct {
	ServerQueryData

	Name           string
	Description    string
	Version        string
	Tags           []string
	Address        string
	Players        []string
	Mods           []FactorioModModel
	ServerResponse []byte
	Spectrum       map[int][]utils.Spectre
}

type FactorioModModel struct {
	Name    string
	Version string
	Hash    string
}

type FactorioServerData struct {
	Address string
	Port    int
}

func (serverData *FactorioServerData) QueryServer() (*FactorioServerDataModel, error) {
	var name, description, version string
	var tags, players []string
	var mods []FactorioModModel

	if serverData.Port == 0 {
		serverData.Port = 34197
	}

	conn, err := net.DialTimeout("udp", serverData.Address+":"+strconv.Itoa(serverData.Port), time.Duration(100)*time.Millisecond)

	if err != nil {
		fmt.Printf("Some error %v", err)
		return nil, err
	}

	ping := []byte{0x02, 0xa8, 0x76, 0x00, 0x00, 0x00, 0x8a, 0x74, 0x39, 0x22, 0x5b, 0x86}

	reader := bufio.NewReader(conn)

	conn.Write(ping)

	response := make([]byte, 32)
	_, err = reader.Read(response)

	reader = bufio.NewReader(conn)

	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return nil, err
	}

	serverMajor, serverMinor, serverPatch := int(response[3]), int(response[4]), int(response[5])

	version = strconv.Itoa(int(response[3])) + "." + strconv.Itoa(int(response[4])) + "." + strconv.Itoa(int(response[5]))

	conn.Write([]byte{0x04, 0xa9, 0x76, 0x39, 0x22, 0x5b, 0x86, response[12], response[13], response[14], response[15], 0x6b, 0x39, 0x34, 0x9d, 0x08, 0x4a, 0x6f, 0x68, 0x6e, 0x20, 0x44, 0x6f, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6e, 0xae, 0x93, 0xd3, 0x01, 0x00, 0x00, 0x00, 0x04, 0x62, 0x61, 0x73, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	parser := utils.NewParser(reader)

	// Initialize
	reader.Peek(1)

	parser.DiscardBytes(8)

	// No idea what this garbage is
	parser.SkipWhile(byte(0))

	// Skip over potential garbage
	parser.ReadString(4)

	if serverMajor == 0 && serverMinor == 15 && serverPatch >= 19 {
		// Skip over potential garbage
		parser.ReadString(4)
	}

	// No idea what this garbage is
	parser.SkipWhile(byte(0))

	// Skip over random byte
	parser.DiscardBytes(1)

	// No idea what this garbage is
	parser.SkipWhile(byte(255))

	// Skip over server player name
	parser.ReadString(1)

	// Offset 2 null bytes
	parser.DiscardBytes(2)

	// Read player count
	var playerCount = parser.ReadInt(1)

	players = make([]string, playerCount)

	// Read all players
	for i := 0; i < playerCount; i++ {
		// Skip player ID?
		parser.DiscardBytes(1)

		// Read player name
		_, players[i] = parser.ReadString(1)

		// Skip null
		parser.DiscardBytes(1)
	}

	// Offset 10 garbage? bytes
	parser.DiscardBytes(10)

	// No idea what this garbage is
	parser.SkipWhile(byte(0))

	// No idea what this garbage is
	parser.SkipWhile(byte(255))

	// Read player count
	var modCount = parser.ReadInt(4)

	mods = make([]FactorioModModel, 0)

	// Read all players
	for i := 0; i < modCount; i++ {
		// Read mod name
		var modLength, modName = parser.ReadString(1)

		if modLength == 0 {
			break
		}

		// Read mod version
		var major = parser.ReadInt(1)
		var minor = parser.ReadInt(1)
		var patch = parser.ReadInt(1)

		// Read mod hash
		var modHash = string(parser.ReadBytes(4))

		mods = append(mods, FactorioModModel{
			Name:    modName,
			Version: strconv.Itoa(major) + "." + strconv.Itoa(minor) + "." + strconv.Itoa(patch),
			Hash:    modHash,
		})
	}

	// Offset 10 garbage? bytes
	parser.DiscardBytes(10)

	// Read name length
	var nameLength = parser.PeekInt(4)

	// TODO Temporary
	if nameLength >= 255 {
		parser.ProcessAllBytes()

		conn.Close()

		return &FactorioServerDataModel{
			ServerQueryData: ServerQueryData{},
			Name:            name,
			Description:     description,
			Version:         version,
			Tags:            tags,
			Address:         serverData.Address + ":" + strconv.Itoa(serverData.Port),
			Players:         players,
			Mods:            mods,
			ServerResponse:  parser.Buffer.Data,
			Spectrum:        parser.Spectrometer(),
		}, errors.New("Could not process packet correctly")
	}

	// Read mod name
	_, name = parser.ReadString(4)

	// Offset 5 garbage? bytes
	parser.DiscardBytes(5)

	// Read mod description
	_, description = parser.ReadString(4)

	if serverMajor == 0 && serverMinor == 15 && serverPatch >= 19 {
		// No idea what this is
		parser.DiscardBytes(4)
	}

	// No idea what this garbage is
	parser.SkipWhile(byte(0))

	_, address := parser.ReadString(4)

	// Read tag count
	var tagCount = parser.ReadInt(4)

	tags = make([]string, tagCount)

	for i := 0; i < int(tagCount); i++ {
		// Read tag
		_, tags[i] = parser.ReadString(4)
	}

	parser.ProcessAllBytes()

	conn.Close()

	return &FactorioServerDataModel{
		ServerQueryData: ServerQueryData{},
		Name:            name,
		Description:     description,
		Version:         version,
		Tags:            tags,
		Address:         address,
		Players:         players,
		Mods:            mods,
		ServerResponse:  parser.Buffer.Data,
		Spectrum:        parser.Spectrometer(),
	}, nil
}
