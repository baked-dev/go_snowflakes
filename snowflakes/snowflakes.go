package snowflakes

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	NodeId     uint16
	Epoch      uint64
	Seq        uint16
	SigningKey string
	mu         sync.Mutex
}

type Snowflake struct {
	FlakeType   string
	Seq         uint16
	Sig         string
	Ts          time.Time
	Data        string
	ParentsData []string
}

func NewClient() *Client {
	return &Client{
		NodeId:     1023,
		Epoch:      1618868000000,
		Seq:        0,
		SigningKey: "",
	}
}

func NewClientWithSigningKey(signingKey string) *Client {
	return &Client{
		NodeId:     1023,
		Epoch:      1618868000000,
		Seq:        0,
		SigningKey: signingKey,
	}
}

func NewClientWithSigningKeyAndNodeId(signingKey string, nodeId uint16) *Client {
	return &Client{
		NodeId:     nodeId,
		Epoch:      1618868000000,
		Seq:        0,
		SigningKey: signingKey,
	}
}

func (c *Client) Gen(flakeType string) (string, error) {
	base := c.GenBase()

	signature, err := c.Sign(flakeType, base, make([]string, 0))
	if err != nil {
		return "", err
	}
	return flakeType + "_" + signature, nil
}

func (c *Client) GenChild(flakeType string, parentFlake string) (string, error) {
	data, _ := c.Read(parentFlake)
	base := c.GenBase()

	signature, err := c.Sign(flakeType, base, append([]string{data.Data}, data.ParentsData...))
	if err != nil {
		return "", err
	}
	return flakeType + "_" + signature, nil
}

func (c *Client) GenBase() string {
	c.mu.Lock()
	seq := c.Seq
	if c.Seq > 4095 {
		c.Seq = 0
	} else {
		c.Seq++
	}
	c.mu.Unlock()
	ts := getTimestamp() - c.Epoch
	tsBin := fmt.Sprintf("%048s", strconv.FormatUint(ts, 2))

	nodeIdBim := fmt.Sprintf("%010s", strconv.FormatUint(uint64(c.NodeId&1023), 2))
	seqBin := fmt.Sprintf("%012s", strconv.FormatUint(uint64(seq), 2))

	finalBin := tsBin + nodeIdBim + seqBin

	num, _ := new(big.Int).SetString(finalBin, 2)

	return num.Text(16)
}

func (c *Client) getSignature(flakeType string, data string, parents []string) (string, [][]string, error) {
	splitParents := [][]string{}
	for _, element2 := range parents {
		splitParent := strings.Split(element2, "")
		splitParents = append(splitParents, splitParent)
	}
	signaturePayload := data
	for _, element := range parents {
		signaturePayload += element
	}
	signaturePayload += c.SigningKey
	signaturePayload += flakeType
	return fmt.Sprintf("%x", sha256.Sum256([]byte(signaturePayload))), splitParents, nil
}

func (c *Client) Sign(flakeType string, data string, parents []string) (string, error) {
	for index, element := range parents {
		if index%2 == 0 {
			parents[index] = Reverse(element)
		}
	}
	signature, splitParents, _ := c.getSignature(flakeType, data, parents)
	splitSignature := strings.Split(signature, "")
	splitData := strings.Split(data, "")
	result := ""
	for idx, el := range splitData {
		result += splitSignature[idx]
		result += el
		for _, el2 := range splitParents {
			result += el2[idx]
		}
	}
	return result, nil
}

func (c *Client) Read(flake string) (*Snowflake, error) {
	regex := regexp.MustCompile("_")
	splitFlake := regex.Split(flake, -1)
	flakeRaw := splitFlake[len(splitFlake)-1]
	flakeType := strings.Join(splitFlake[:len(splitFlake)-1], "_")
	splitRaw := strings.Split(flakeRaw, "")
	unmangledData := []string{}
	dataAmount := len(splitRaw) / 14
	for i := 0; i < dataAmount; i++ {
		unmangledData = append(unmangledData, "")
	}
	for idx, el := range splitRaw {
		index := idx % dataAmount
		unmangledData[index] += el
	}
	flakeSignature := unmangledData[0]
	unmangledData = unmangledData[1:]
	flakeData := unmangledData[0]
	unmangledData = unmangledData[1:]
	for index, element := range unmangledData {
		if index%2 == 0 {
			unmangledData[index] = Reverse(element)
		}
	}
	ts, seq, _ := parseData(flakeData)
	return &Snowflake{
		FlakeType:   flakeType,
		Sig:         flakeSignature,
		Data:        flakeData,
		ParentsData: unmangledData,
		Seq:         seq,
		Ts:          time.Unix(0, (ts+int64(c.Epoch))*int64(time.Millisecond)),
	}, nil
}

func (c *Client) GenParent(flake string, parentType string) (string, error) {
	parsed, err := c.Read(flake)
	if err != nil {
		return "", err
	}
	parentData := parsed.ParentsData[0]
	parsed.ParentsData = parsed.ParentsData[1:]
	tmp := make([]string, len(parsed.ParentsData))
	copy(tmp, parsed.ParentsData)
	sig, err := c.Sign(parentType, parentData, tmp)
	if err != nil {
		return "", err
	}
	return parentType + "_" + sig, nil
}

func (c *Client) Verify(flake string) (*Snowflake, error) {
	data, err := c.Read(flake)
	if err != nil {
		return nil, err
	}
	tmp := make([]string, len(data.ParentsData))
	copy(tmp, data.ParentsData)
	signature, _, err := c.getSignature(data.FlakeType, data.Data, tmp)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(signature, data.Sig) {
		return data, nil
	} else {
		return data, errors.New("couldnt verify signature")
	}
}

func parseData(data string) (int64, uint16, error) {
	num, _ := new(big.Int).SetString(data, 16)
	bin := fmt.Sprintf("%070s", num.Text(2))
	splitBin := strings.Split(bin, "")
	binTs := strings.Join(splitBin[0:48], "")
	binSeq := strings.Join(splitBin[58:], "")
	intTs, _ := strconv.ParseUint(binTs, 2, 64)
	intSeq, _ := strconv.ParseUint(binSeq, 2, 16)
	return int64(intTs), uint16(intSeq), nil
}

func getTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
