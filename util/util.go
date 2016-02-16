package util

import (
	"encoding/hex"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/decred/dcrd/blockchain"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
)

var PowLimit = chaincfg.MainNetParams.PowLimit

const (
// uint256Size is the number of bytes needed to represent an unsigned
// 256-bit integer.
	uint256Size = 32
)

func Random() string {
	min := int64(100000000000000)
	max := int64(999999999999999)
	n := rand.Int63n(max-min+1) + min
	return strconv.FormatInt(n, 10)
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MakeTargetHex(minerDifficulty int64) string {
	difficulty := bigToLEUint256(MakeTarget(minerDifficulty))
	return hex.EncodeToString(difficulty[:])
}

// bigToLEUint256 returns the passed big integer as an unsigned 256-bit integer
// encoded as little-endian bytes.  Numbers which are larger than the max
// unsigned 256-bit integer are truncated.
func bigToLEUint256(n *big.Int) [uint256Size]byte {
	// Pad or truncate the big-endian big int to correct number of bytes.
	nBytes := n.Bytes()
	nlen := len(nBytes)
	pad := 0
	start := 0
	if nlen <= uint256Size {
		pad = uint256Size - nlen
	} else {
		start = nlen - uint256Size
	}
	var buf [uint256Size]byte
	copy(buf[pad:], nBytes[start:])

	// Reverse the bytes to little endian and return them.
	for i := 0; i < uint256Size/2; i++ {
		buf[i], buf[uint256Size-1-i] = buf[uint256Size-1-i], buf[i]
	}
	return buf
}

func MakeTarget(minerDifficulty int64) *big.Int {
	targetDifficulty := big.NewInt(minerDifficulty)
	return targetDifficulty.Div(PowLimit, targetDifficulty)
}

//func DiffToBig(targetDifficulty big.Int) *big.Int {
//	return targetDifficulty.Div(PowLimit, targetDifficulty)
//}

func BigToDiff(n *big.Int) int64 {
	return n.Div(PowLimit, n).Int64()
}

/* convert target to big integer
"000000003fffffffffffffffffffffffffffffffffffffffffffffffffffffff"
*/
func TargetHexToDiff(targetHex string) *big.Int {
	// hash, err := chainhash.NewHashFromStr(targetHex)
	hashbytes, err := hex.DecodeString(targetHex)
	if err != nil {
		return big.NewInt(0)
	}
	// return blockchain.ShaHashToBig(hash)
	return new(big.Int).SetBytes(hashbytes[:])
	/*
	targetBytes := common.FromHex(targetHex)
	return new(big.Int).Div(pow256, common.BytesToBig(targetBytes))
	*/
}


/* convert target come from getwork to big integer
   "ffffffffffffffffffffffffffffffffffffffffffffffffffffff3f00000000"
 */
func TargetStrToDiff(targetHex string) *big.Int {
	hashbytes, err := hex.DecodeString(targetHex)
	if err != nil {
		return big.NewInt(0)
	}
	hash, err := chainhash.NewHash(hashbytes)
	if err != nil {
		return big.NewInt(0)
	}
	targetdiff := blockchain.ShaHashToBig(hash)
	if targetdiff.Cmp( big.NewInt(0)) != 0 {
		targetdiff.Div(PowLimit, targetdiff)
	}
	return targetdiff
}

// getDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func GetDifficultyRatio(bits uint32) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number.  Note this is not the same as the the
	// proof of work limit directly because the block difficulty is encoded
	// in a block with the compact form which loses precision.
	max := PowLimit // blockchain.CompactToBig(chaincfg.TestNetParams.PowLimitBits)
	target := blockchain.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(max, target)
	outString := difficulty.FloatString(8)
	diff, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		//rpcsLog.Errorf("Cannot get difficulty: %v", err)
		return 0
	}
	return diff
}
