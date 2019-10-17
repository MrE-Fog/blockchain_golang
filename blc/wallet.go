package block

import (
	"bytes"
	"encoding/gob"
	"github.com/corgi-kx/blockchain_golang/database"
	log "github.com/corgi-kx/logcustom"
)

type addressList [][]byte

func (a *addressList) serliazle() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

type wallets struct {
	Wallets map[string]*bitcoinKeys
}

func NewWallets(bd *database.BlockchainDB) *wallets {
	w := &wallets{make(map[string]*bitcoinKeys)}
	if database.IsBucketExist(bd, database.AddrBucket) {
		addressList := GetAllAddress(bd)
		if addressList == nil {
			return w
		}
		for _, v := range *addressList {
			keys := Deserialize(bd.View(v, database.AddrBucket), &bitcoinKeys{})
			w.Wallets[string(v)] = keys.(*bitcoinKeys)
		}
		return w
	}
	return w
}

func (w *wallets) GenerateWallet(bd *database.BlockchainDB,keys func([]string) *bitcoinKeys,s []string) (address, privKey, mnemonicWord string) {
	bitcoinKeys := keys(s)
	if bitcoinKeys == nil {
		log.Fatal("创建钱包失败，检查助记词是否符合创建规则！")
	}
	privKey = bitcoinKeys.GetPrivateKey()
	addressByte := bitcoinKeys.getAddress()
	w.storage(addressByte, bitcoinKeys, bd)
	//将地址存入实例
	address = string(addressByte)
	for _, v := range bitcoinKeys.MnemonicWord {
		mnemonicWord += v + " "
	}
	//w.Wallets[address] = bitcoinKeys
	return
}

func (w *wallets) storage(address []byte, keys *bitcoinKeys, bd *database.BlockchainDB) {
	b:=bd.View(address, database.AddrBucket)
	if b != nil {
		log.Warn("钱包早已存在于数据库中！")
		return
	}
	//将公私钥以地址为键 存入数据库
	bd.Put(address, keys.serliazle(), database.AddrBucket)
	//将地址存入地址导航
	listBytes := bd.View([]byte(addrListMapping), database.AddrBucket)
	if listBytes == nil {
		a := addressList{address}
		bd.Put([]byte(addrListMapping), a.serliazle(), database.AddrBucket)
	} else {
		a := Deserialize(listBytes, &addressList{})
		addressList := a.(*addressList)
		*addressList = append(*addressList, address)
		bd.Put([]byte(addrListMapping), addressList.serliazle(), database.AddrBucket)
	}
}

//获取全部地址信息
func GetAllAddress(bd *database.BlockchainDB) *addressList {
	listBytes := bd.View([]byte(addrListMapping), database.AddrBucket)
	if listBytes == nil {
		return nil
	}
	a := Deserialize(listBytes, &addressList{})
	addressList := a.(*addressList)
	return addressList
}

