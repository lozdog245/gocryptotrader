package currency

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
)

// Bitmasks const for currency rolls
const (
	Unset Role = 0
	Fiat  Role = 1 << (iota - 1)
	Cryptocurrency
	Token
	Contract

	UnsetRollString      = "roleUnset"
	FiatCurrencyString   = "fiatCurrency"
	CryptocurrencyString = "cryptocurrency"
	TokenString          = "token"
	ContractString       = "contract"
)

// Role defines a bitmask for the full currency rolls either; fiat,
// cryptocurrency, token, or contract
type Role uint8

func (r Role) String() string {
	switch r {
	case Unset:
		return UnsetRollString
	case Fiat:
		return FiatCurrencyString
	case Cryptocurrency:
		return CryptocurrencyString
	case Token:
		return TokenString
	case Contract:
		return ContractString
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON conforms Roll to the marshaler interface
func (r Role) MarshalJSON() ([]byte, error) {
	return common.JSONEncode(r.String())
}

// UnmarshalJSON conforms Roll to the unmarshaller interface
func (r *Role) UnmarshalJSON(d []byte) error {
	var incoming string
	err := common.JSONDecode(d, &incoming)
	if err != nil {
		return err
	}

	switch incoming {
	case UnsetRollString:
		*r = Unset
	case FiatCurrencyString:
		*r = Fiat
	case CryptocurrencyString:
		*r = Cryptocurrency
	case TokenString:
		*r = Token
	case ContractString:
		*r = Contract
	default:
		return fmt.Errorf("unmarshal error role type %s unsupported for currency",
			incoming)
	}
	return nil
}

// BaseCodes defines a basket of bare currency codes
type BaseCodes struct {
	Items          []*Item
	LastMainUpdate time.Time
	mtx            sync.Mutex
}

// HasData returns true if the type contains data
func (b *BaseCodes) HasData() bool {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return len(b.Items) != 0
}

// GetFullCurrencyData returns a type that is read to dump to file
func (b *BaseCodes) GetFullCurrencyData() (File, error) {
	var file File
	for _, i := range b.Items {
		switch i.Role {
		case Unset:
			file.UnsetCurrency = append(file.UnsetCurrency, *i)
		case Fiat:
			file.FiatCurrency = append(file.FiatCurrency, *i)
		case Cryptocurrency:
			file.Cryptocurrency = append(file.Cryptocurrency, *i)
		case Token:
			file.Token = append(file.Token, *i)
		case Contract:
			file.Contracts = append(file.Contracts, *i)
		default:
			return file, errors.New("roll undefined")
		}
	}

	file.LastMainUpdate = b.LastMainUpdate
	return file, nil
}

// GetCurrencies gets the full currency list from the base code type available
// from the currency system
func (b *BaseCodes) GetCurrencies() Currencies {
	var currencies Currencies
	b.mtx.Lock()
	for i := range b.Items {
		currencies = append(currencies, Code{
			Item: b.Items[i],
		})
	}
	b.mtx.Unlock()
	return currencies
}

// UpdateCryptocurrency updates or registers a cryptocurrency
func (b *BaseCodes) UpdateCryptocurrency(fullName, symbol string, id int) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol != symbol {
			continue
		}
		if b.Items[i].Role != Unset {
			if b.Items[i].Role != Cryptocurrency {
				if b.Items[i].FullName != "" {
					if b.Items[i].FullName != fullName {
						// multiple symbols found, break this and add the
						// full context - this most likely won't occur for
						// fiat but could occur for contracts.
						break
					}
				}
				return fmt.Errorf("role already defined in cryptocurrency %s as [%s]",
					b.Items[i].Symbol,
					b.Items[i].Role)
			}
			b.Items[i].FullName = fullName
			b.Items[i].ID = id
			return nil
		}

		b.Items[i].Role = Cryptocurrency
		b.Items[i].FullName = fullName
		b.Items[i].ID = id
		return nil
	}

	b.Items = append(b.Items, &Item{
		FullName: fullName,
		Symbol:   symbol,
		ID:       id,
		Role:     Cryptocurrency,
	})
	return nil
}

// UpdateFiatCurrency updates or registers a fiat currency
func (b *BaseCodes) UpdateFiatCurrency(fullName, symbol string, id int) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol != symbol {
			continue
		}

		if b.Items[i].Role != Unset {
			if b.Items[i].Role != Fiat {
				return fmt.Errorf("role already defined in fiat currency %s as [%s]",
					b.Items[i].Symbol,
					b.Items[i].Role)
			}
			b.Items[i].FullName = fullName
			b.Items[i].ID = id
			return nil
		}

		b.Items[i].Role = Fiat
		b.Items[i].FullName = fullName
		b.Items[i].ID = id
		return nil
	}

	b.Items = append(b.Items, &Item{
		FullName: fullName,
		Symbol:   symbol,
		ID:       id,
		Role:     Fiat,
	})
	return nil
}

// UpdateToken updates or registers a token
func (b *BaseCodes) UpdateToken(fullName, symbol, assocBlockchain string, id int) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol != symbol {
			continue
		}

		if b.Items[i].Role != Unset {
			if b.Items[i].Role != Token {
				if b.Items[i].FullName != "" {
					if b.Items[i].FullName != fullName {
						// multiple symbols found, break this and add the
						// full context - this most likely won't occur for
						// fiat but could occur for contracts.
						break
					}
				}
				return fmt.Errorf("role already defined in token %s as [%s]",
					b.Items[i].Symbol,
					b.Items[i].Role)
			}
			b.Items[i].FullName = fullName
			b.Items[i].ID = id
			b.Items[i].AssocChain = assocBlockchain
			return nil
		}

		b.Items[i].Role = Token
		b.Items[i].FullName = fullName
		b.Items[i].ID = id
		b.Items[i].AssocChain = assocBlockchain
		return nil
	}

	b.Items = append(b.Items, &Item{
		FullName:   fullName,
		Symbol:     symbol,
		ID:         id,
		Role:       Token,
		AssocChain: assocBlockchain,
	})
	return nil
}

// UpdateContract updates or registers a contract
func (b *BaseCodes) UpdateContract(fullName, symbol, assocExchange string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol != symbol {
			continue
		}

		if b.Items[i].Role != Unset {
			if b.Items[i].Role != Contract {
				return fmt.Errorf("role already defined in contract %s as [%s]",
					b.Items[i].Symbol,
					b.Items[i].Role)
			}
			b.Items[i].FullName = fullName
			if !common.StringDataContains(b.Items[i].AssocExchange, assocExchange) {
				b.Items[i].AssocExchange = append(b.Items[i].AssocExchange,
					assocExchange)
			}
			return nil
		}

		b.Items[i].Role = Contract
		b.Items[i].FullName = fullName
		if !common.StringDataContains(b.Items[i].AssocExchange, assocExchange) {
			b.Items[i].AssocExchange = append(b.Items[i].AssocExchange,
				assocExchange)
		}
		return nil
	}

	b.Items = append(b.Items, &Item{
		FullName:      fullName,
		Symbol:        symbol,
		Role:          Contract,
		AssocExchange: []string{assocExchange},
	})
	return nil
}

// Register registers a currency from a string and returns a currency code
func (b *BaseCodes) Register(c string) Code {
	NewUpperCode := common.StringToUpper(c)
	format := common.StringContains(c, NewUpperCode)

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol == NewUpperCode {
			return Code{
				Item:      b.Items[i],
				UpperCase: format,
			}
		}
	}

	newItem := Item{Symbol: NewUpperCode}
	newCode := Code{
		Item:      &newItem,
		UpperCase: format,
	}

	b.Items = append(b.Items, newCode.Item)
	return newCode
}

// RegisterFiat registers a fiat currency from a string and returns a currency
// code
func (b *BaseCodes) RegisterFiat(c string) (Code, error) {
	c = common.StringToUpper(c)

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol == c {
			if b.Items[i].Role != Unset {
				if b.Items[i].Role != Fiat {
					return Code{}, fmt.Errorf("register fiat error role already defined in fiat %s as [%s]",
						b.Items[i].Symbol,
						b.Items[i].Role)
				}
				return Code{Item: b.Items[i], UpperCase: true}, nil
			}
			b.Items[i].Role = Fiat
			return Code{Item: b.Items[i], UpperCase: true}, nil
		}
	}

	item := &Item{Symbol: c, Role: Fiat}
	b.Items = append(b.Items, item)

	return Code{Item: item, UpperCase: true}, nil
}

// LoadItem sets item data
func (b *BaseCodes) LoadItem(item *Item) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := range b.Items {
		if b.Items[i].Symbol == item.Symbol {
			if b.Items[i].Role == Unset {
				b.Items[i].AssocChain = item.AssocChain
				b.Items[i].AssocExchange = item.AssocExchange
				b.Items[i].ID = item.ID
				b.Items[i].Role = item.Role
				b.Items[i].FullName = item.FullName
				return nil
			}

			if b.Items[i].FullName != "" {
				if b.Items[i].FullName == item.FullName {
					b.Items[i].AssocChain = item.AssocChain
					b.Items[i].AssocExchange = item.AssocExchange
					b.Items[i].ID = item.ID
					b.Items[i].Role = item.Role
					return nil
				}
				break
			}

			if b.Items[i].ID == item.ID {
				b.Items[i].AssocChain = item.AssocChain
				b.Items[i].AssocExchange = item.AssocExchange
				b.Items[i].FullName = item.FullName
				b.Items[i].ID = item.ID
				b.Items[i].Role = item.Role
				return nil
			}

			return fmt.Errorf("currency %s not found in currencycode list",
				item.Symbol)
		}
	}

	b.Items = append(b.Items, item)
	return nil
}

// NewCode returns a new currency registered code
func NewCode(c string) Code {
	return storage.ValidateCode(c)
}

// Code defines an ISO 4217 fiat currency or unofficial cryptocurrency code
// string
type Code struct {
	Item      *Item
	UpperCase bool
}

// Item defines a sub type containing the main attributes of a designated
// currency code pointer
type Item struct {
	ID            int      `json:"id"`
	FullName      string   `json:"fullName"`
	Symbol        string   `json:"symbol"`
	Role          Role     `json:"role"`
	AssocChain    string   `json:"associatedBlockchain"`
	AssocExchange []string `json:"associatedExchanges"`
}

// String conforms to the stringer interface
func (i *Item) String() string {
	return i.FullName
}

// String converts the code to string
func (c Code) String() string {
	if c.Item == nil {
		return ""
	}

	if c.UpperCase {
		return c.Item.Symbol
	}
	return common.StringToLower(c.Item.Symbol)
}

// Lower converts the code to lowercase formatting
func (c Code) Lower() Code {
	c.UpperCase = false
	return c
}

// Upper converts the code to uppercase formatting
func (c Code) Upper() Code {
	c.UpperCase = true
	return c
}

// UnmarshalJSON comforms type to the umarshaler interface
func (c *Code) UnmarshalJSON(d []byte) error {
	var newcode string
	err := common.JSONDecode(d, &newcode)
	if err != nil {
		return err
	}
	*c = NewCode(newcode)
	return nil
}

// MarshalJSON conforms type to the marshaler interface
func (c Code) MarshalJSON() ([]byte, error) {
	if c.Item == nil {
		return common.JSONEncode("")
	}
	return common.JSONEncode(c.String())
}

// IsEmpty returns true if the code is empty
func (c Code) IsEmpty() bool {
	if c.Item == nil {
		return true
	}
	return c.Item.Symbol == ""
}

// Match returns if the code supplied is the same as the corresponding code
func (c Code) Match(check Code) bool {
	return c.Item == check.Item
}

// IsDefaultFiatCurrency checks if the currency passed in matches the default
// fiat currency
func (c Code) IsDefaultFiatCurrency() bool {
	return storage.IsDefaultCurrency(c)
}

// IsDefaultCryptocurrency checks if the currency passed in matches the default
// cryptocurrency
func (c Code) IsDefaultCryptocurrency() bool {
	return storage.IsDefaultCryptocurrency(c)
}

// IsFiatCurrency checks if the currency passed is an enabled fiat currency
func (c Code) IsFiatCurrency() bool {
	return storage.IsFiatCurrency(c)
}

// IsCryptocurrency checks if the currency passed is an enabled CRYPTO currency.
func (c Code) IsCryptocurrency() bool {
	return storage.IsCryptocurrency(c)
}

// Const declarations for individual currencies/tokens/fiat
// An ever growing list. Cares not for equivalence, just is
var (
	BTC        = NewCode("BTC")
	LTC        = NewCode("LTC")
	ETH        = NewCode("ETH")
	XRP        = NewCode("XRP")
	BCH        = NewCode("BCH")
	EOS        = NewCode("EOS")
	XLM        = NewCode("XLM")
	USDT       = NewCode("USDT")
	ADA        = NewCode("ADA")
	XMR        = NewCode("XMR")
	TRX        = NewCode("TRX")
	MIOTA      = NewCode("MIOTA")
	DASH       = NewCode("DASH")
	BNB        = NewCode("BNB")
	NEO        = NewCode("NEO")
	ETC        = NewCode("ETC")
	XEM        = NewCode("XEM")
	XTZ        = NewCode("XTZ")
	VET        = NewCode("VET")
	DOGE       = NewCode("DOGE")
	ZEC        = NewCode("ZEC")
	OMG        = NewCode("OMG")
	BTG        = NewCode("BTG")
	MKR        = NewCode("MKR")
	BCN        = NewCode("BCN")
	ONT        = NewCode("ONT")
	ZRX        = NewCode("ZRX")
	LSK        = NewCode("LSK")
	DCR        = NewCode("DCR")
	QTUM       = NewCode("QTUM")
	BCD        = NewCode("BCD")
	BTS        = NewCode("BTS")
	NANO       = NewCode("NANO")
	ZIL        = NewCode("ZIL")
	SC         = NewCode("SC")
	DGB        = NewCode("DGB")
	ICX        = NewCode("ICX")
	STEEM      = NewCode("STEEM")
	AE         = NewCode("AE")
	XVG        = NewCode("XVG")
	WAVES      = NewCode("WAVES")
	NPXS       = NewCode("NPXS")
	ETN        = NewCode("ETN")
	BTM        = NewCode("BTM")
	BAT        = NewCode("BAT")
	ETP        = NewCode("ETP")
	HOT        = NewCode("HOT")
	STRAT      = NewCode("STRAT") // nolint: misspell
	GNT        = NewCode("GNT")
	REP        = NewCode("REP")
	SNT        = NewCode("SNT")
	PPT        = NewCode("PPT")
	KMD        = NewCode("KMD")
	TUSD       = NewCode("TUSD")
	CNX        = NewCode("CNX")
	LINK       = NewCode("LINK")
	WTC        = NewCode("WTC")
	ARDR       = NewCode("ARDR")
	WAN        = NewCode("WAN")
	MITH       = NewCode("MITH")
	RDD        = NewCode("RDD")
	IOST       = NewCode("IOST")
	IOT        = NewCode("IOT")
	KCS        = NewCode("KCS")
	MAID       = NewCode("MAID")
	XET        = NewCode("XET")
	MOAC       = NewCode("MOAC")
	HC         = NewCode("HC")
	AION       = NewCode("AION")
	AOA        = NewCode("AOA")
	HT         = NewCode("HT")
	ELF        = NewCode("ELF")
	LRC        = NewCode("LRC")
	BNT        = NewCode("BNT")
	CMT        = NewCode("CMT")
	DGD        = NewCode("DGD")
	DCN        = NewCode("DCN")
	FUN        = NewCode("FUN")
	GXS        = NewCode("GXS")
	DROP       = NewCode("DROP")
	MANA       = NewCode("MANA")
	PAY        = NewCode("PAY")
	MCO        = NewCode("MCO")
	THETA      = NewCode("THETA")
	NXT        = NewCode("NXT")
	NOAH       = NewCode("NOAH")
	LOOM       = NewCode("LOOM")
	POWR       = NewCode("POWR")
	WAX        = NewCode("WAX")
	ELA        = NewCode("ELA")
	PIVX       = NewCode("PIVX")
	XIN        = NewCode("XIN")
	DAI        = NewCode("DAI")
	BTCP       = NewCode("BTCP")
	NEXO       = NewCode("NEXO")
	XBT        = NewCode("XBT")
	SAN        = NewCode("SAN")
	GAS        = NewCode("GAS")
	BCC        = NewCode("BCC")
	HCC        = NewCode("HCC")
	OAX        = NewCode("OAX")
	DNT        = NewCode("DNT")
	ICN        = NewCode("ICN")
	LLT        = NewCode("LLT")
	YOYO       = NewCode("YOYO")
	SNGLS      = NewCode("SNGLS")
	BQX        = NewCode("BQX")
	KNC        = NewCode("KNC")
	SNM        = NewCode("SNM")
	CTR        = NewCode("CTR")
	SALT       = NewCode("SALT")
	MDA        = NewCode("MDA")
	IOTA       = NewCode("IOTA")
	SUB        = NewCode("SUB")
	MTL        = NewCode("MTL")
	MTH        = NewCode("MTH")
	ENG        = NewCode("ENG")
	AST        = NewCode("AST")
	CLN        = NewCode("CLN")
	EDG        = NewCode("EDG")
	FIRST      = NewCode("1ST")
	GOLOS      = NewCode("GOLOS")
	ANT        = NewCode("ANT")
	GBG        = NewCode("GBG")
	HMQ        = NewCode("HMQ")
	INCNT      = NewCode("INCNT")
	ACE        = NewCode("ACE")
	ACT        = NewCode("ACT")
	AAC        = NewCode("AAC")
	AIDOC      = NewCode("AIDOC")
	SOC        = NewCode("SOC")
	ATL        = NewCode("ATL")
	AVT        = NewCode("AVT")
	BKX        = NewCode("BKX")
	BEC        = NewCode("BEC")
	VEE        = NewCode("VEE")
	PTOY       = NewCode("PTOY")
	CAG        = NewCode("CAG")
	CIC        = NewCode("CIC")
	CBT        = NewCode("CBT")
	CAN        = NewCode("CAN")
	DAT        = NewCode("DAT")
	DNA        = NewCode("DNA")
	INT        = NewCode("INT")
	IPC        = NewCode("IPC")
	ILA        = NewCode("ILA")
	LIGHT      = NewCode("LIGHT")
	MAG        = NewCode("MAG")
	AMM        = NewCode("AMM")
	MOF        = NewCode("MOF")
	MGC        = NewCode("MGC")
	OF         = NewCode("OF")
	LA         = NewCode("LA")
	LEV        = NewCode("LEV")
	NGC        = NewCode("NGC")
	OKB        = NewCode("OKB")
	MOT        = NewCode("MOT")
	PRA        = NewCode("PRA")
	R          = NewCode("R")
	SSC        = NewCode("SSC")
	SHOW       = NewCode("SHOW")
	SPF        = NewCode("SPF")
	SNC        = NewCode("SNC")
	SWFTC      = NewCode("SWFTC")
	TRA        = NewCode("TRA")
	TOPC       = NewCode("TOPC")
	TRIO       = NewCode("TRIO")
	QVT        = NewCode("QVT")
	UCT        = NewCode("UCT")
	UKG        = NewCode("UKG")
	UTK        = NewCode("UTK")
	VIU        = NewCode("VIU")
	WFEE       = NewCode("WFEE")
	WRC        = NewCode("WRC")
	UGC        = NewCode("UGC")
	YEE        = NewCode("YEE")
	YOYOW      = NewCode("YOYOW")
	ZIP        = NewCode("ZIP")
	READ       = NewCode("READ")
	RCT        = NewCode("RCT")
	REF        = NewCode("REF")
	XUC        = NewCode("XUC")
	FAIR       = NewCode("FAIR")
	GSC        = NewCode("GSC")
	HMC        = NewCode("HMC")
	PLU        = NewCode("PLU")
	PRO        = NewCode("PRO")
	QRL        = NewCode("QRL")
	REN        = NewCode("REN")
	ROUND      = NewCode("ROUND")
	SRN        = NewCode("SRN")
	XID        = NewCode("XID")
	SBD        = NewCode("SBD")
	TAAS       = NewCode("TAAS")
	TKN        = NewCode("TKN")
	VEN        = NewCode("VEN")
	VSL        = NewCode("VSL")
	TRST       = NewCode("TRST")
	XXX        = NewCode("XXX")
	IND        = NewCode("IND")
	LDC        = NewCode("LDC")
	GUP        = NewCode("GUP")
	MGO        = NewCode("MGO")
	MYST       = NewCode("MYST")
	NEU        = NewCode("NEU")
	NET        = NewCode("NET")
	BMC        = NewCode("BMC")
	BCAP       = NewCode("BCAP")
	TIME       = NewCode("TIME")
	CFI        = NewCode("CFI")
	EVX        = NewCode("EVX")
	REQ        = NewCode("REQ")
	VIB        = NewCode("VIB")
	ARK        = NewCode("ARK")
	MOD        = NewCode("MOD")
	ENJ        = NewCode("ENJ")
	STORJ      = NewCode("STORJ")
	RCN        = NewCode("RCN")
	NULS       = NewCode("NULS")
	RDN        = NewCode("RDN")
	DLT        = NewCode("DLT")
	AMB        = NewCode("AMB")
	BCPT       = NewCode("BCPT")
	ARN        = NewCode("ARN")
	GVT        = NewCode("GVT")
	CDT        = NewCode("CDT")
	POE        = NewCode("POE")
	QSP        = NewCode("QSP")
	XZC        = NewCode("XZC")
	TNT        = NewCode("TNT")
	FUEL       = NewCode("FUEL")
	ADX        = NewCode("ADX")
	CND        = NewCode("CND")
	LEND       = NewCode("LEND")
	WABI       = NewCode("WABI")
	SBTC       = NewCode("SBTC")
	BCX        = NewCode("BCX")
	TNB        = NewCode("TNB")
	GTO        = NewCode("GTO")
	OST        = NewCode("OST")
	CVC        = NewCode("CVC")
	DATA       = NewCode("DATA")
	ETF        = NewCode("ETF")
	BRD        = NewCode("BRD")
	NEBL       = NewCode("NEBL")
	VIBE       = NewCode("VIBE")
	LUN        = NewCode("LUN")
	CHAT       = NewCode("CHAT")
	RLC        = NewCode("RLC")
	INS        = NewCode("INS")
	VIA        = NewCode("VIA")
	BLZ        = NewCode("BLZ")
	SYS        = NewCode("SYS")
	NCASH      = NewCode("NCASH")
	POA        = NewCode("POA")
	STORM      = NewCode("STORM")
	WPR        = NewCode("WPR")
	QLC        = NewCode("QLC")
	GRS        = NewCode("GRS")
	CLOAK      = NewCode("CLOAK")
	ZEN        = NewCode("ZEN")
	SKY        = NewCode("SKY")
	IOTX       = NewCode("IOTX")
	QKC        = NewCode("QKC")
	AGI        = NewCode("AGI")
	NXS        = NewCode("NXS")
	EON        = NewCode("EON")
	KEY        = NewCode("KEY")
	NAS        = NewCode("NAS")
	ADD        = NewCode("ADD")
	MEETONE    = NewCode("MEETONE")
	ATD        = NewCode("ATD")
	MFT        = NewCode("MFT")
	EOP        = NewCode("EOP")
	DENT       = NewCode("DENT")
	IQ         = NewCode("IQ")
	DOCK       = NewCode("DOCK")
	POLY       = NewCode("POLY")
	VTHO       = NewCode("VTHO")
	ONG        = NewCode("ONG")
	PHX        = NewCode("PHX")
	GO         = NewCode("GO")
	PAX        = NewCode("PAX")
	EDO        = NewCode("EDO")
	WINGS      = NewCode("WINGS")
	NAV        = NewCode("NAV")
	TRIG       = NewCode("TRIG")
	APPC       = NewCode("APPC")
	KRW        = NewCode("KRW")
	HSR        = NewCode("HSR")
	ETHOS      = NewCode("ETHOS")
	CTXC       = NewCode("CTXC")
	ITC        = NewCode("ITC")
	TRUE       = NewCode("TRUE")
	ABT        = NewCode("ABT")
	RNT        = NewCode("RNT")
	PLY        = NewCode("PLY")
	PST        = NewCode("PST")
	KICK       = NewCode("KICK")
	BTCZ       = NewCode("BTCZ")
	DXT        = NewCode("DXT")
	STQ        = NewCode("STQ")
	INK        = NewCode("INK")
	HBZ        = NewCode("HBZ")
	USDT_ETH   = NewCode("USDT_ETH") // nolint: stylecheck, golint
	QTUM_ETH   = NewCode("QTUM_ETH") // nolint: stylecheck
	BTM_ETH    = NewCode("BTM_ETH")  // nolint: stylecheck, golint
	FIL        = NewCode("FIL")
	STX        = NewCode("STX")
	BOT        = NewCode("BOT")
	VERI       = NewCode("VERI")
	ZSC        = NewCode("ZSC")
	QBT        = NewCode("QBT")
	MED        = NewCode("MED")
	QASH       = NewCode("QASH")
	MDS        = NewCode("MDS")
	GOD        = NewCode("GOD")
	SMT        = NewCode("SMT")
	BTF        = NewCode("BTF")
	NAS_ETH    = NewCode("NAS_ETH") // nolint: stylecheck, golint
	TSL        = NewCode("TSL")
	BIFI       = NewCode("BIFI")
	BNTY       = NewCode("BNTY")
	DRGN       = NewCode("DRGN")
	GTC        = NewCode("GTC")
	MDT        = NewCode("MDT")
	QUN        = NewCode("QUN")
	GNX        = NewCode("GNX")
	DDD        = NewCode("DDD")
	BTO        = NewCode("BTO")
	TIO        = NewCode("TIO")
	OCN        = NewCode("OCN")
	RUFF       = NewCode("RUFF")
	TNC        = NewCode("TNC")
	SNET       = NewCode("SNET")
	COFI       = NewCode("COFI")
	ZPT        = NewCode("ZPT")
	JNT        = NewCode("JNT")
	MTN        = NewCode("MTN")
	GEM        = NewCode("GEM")
	DADI       = NewCode("DADI")
	RFR        = NewCode("RFR")
	MOBI       = NewCode("MOBI")
	LEDU       = NewCode("LEDU")
	DBC        = NewCode("DBC")
	MKR_OLD    = NewCode("MKR_OLD") // nolint: stylecheck, golint
	DPY        = NewCode("DPY")
	BCDN       = NewCode("BCDN")
	EOSDAC     = NewCode("EOSDAC") // nolint: stylecheck
	TIPS       = NewCode("TIPS")
	XMC        = NewCode("XMC")
	PPS        = NewCode("PPS")
	BOE        = NewCode("BOE")
	MEDX       = NewCode("MEDX")
	SMT_ETH    = NewCode("SMT_ETH") // nolint: stylecheck, golint
	CS         = NewCode("CS")
	MAN        = NewCode("MAN")
	REM        = NewCode("REM")
	LYM        = NewCode("LYM")
	INSTAR     = NewCode("INSTAR") // nolint: stylecheck
	BFT        = NewCode("BFT")
	IHT        = NewCode("IHT")
	SENC       = NewCode("SENC")
	TOMO       = NewCode("TOMO")
	ELEC       = NewCode("ELEC")
	SHIP       = NewCode("SHIP")
	TFD        = NewCode("TFD")
	HAV        = NewCode("HAV")
	HUR        = NewCode("HUR")
	LST        = NewCode("LST")
	LINO       = NewCode("LINO")
	SWTH       = NewCode("SWTH")
	NKN        = NewCode("NKN")
	SOUL       = NewCode("SOUL")
	GALA_NEO   = NewCode("GALA_NEO") // nolint: stylecheck, golint
	LRN        = NewCode("LRN")
	GSE        = NewCode("GSE")
	RATING     = NewCode("RATING")
	HSC        = NewCode("HSC")
	HIT        = NewCode("HIT")
	DX         = NewCode("DX")
	BXC        = NewCode("BXC")
	GARD       = NewCode("GARD")
	FTI        = NewCode("FTI")
	SOP        = NewCode("SOP")
	LEMO       = NewCode("LEMO")
	RED        = NewCode("RED")
	LBA        = NewCode("LBA")
	KAN        = NewCode("KAN")
	OPEN       = NewCode("OPEN")
	SKM        = NewCode("SKM")
	NBAI       = NewCode("NBAI")
	UPP        = NewCode("UPP")
	ATMI       = NewCode("ATMI")
	TMT        = NewCode("TMT")
	BBK        = NewCode("BBK")
	EDR        = NewCode("EDR")
	MET        = NewCode("MET")
	TCT        = NewCode("TCT")
	EXC        = NewCode("EXC")
	CNC        = NewCode("CNC")
	TIX        = NewCode("TIX")
	XTC        = NewCode("XTC")
	BU         = NewCode("BU")
	GNO        = NewCode("GNO")
	MLN        = NewCode("MLN")
	XBC        = NewCode("XBC")
	BTCD       = NewCode("BTCD")
	BURST      = NewCode("BURST")
	CLAM       = NewCode("CLAM")
	XCP        = NewCode("XCP")
	EMC2       = NewCode("EMC2")
	EXP        = NewCode("EXP")
	FCT        = NewCode("FCT")
	GAME       = NewCode("GAME")
	GRC        = NewCode("GRC")
	HUC        = NewCode("HUC")
	LBC        = NewCode("LBC")
	NMC        = NewCode("NMC")
	NEOS       = NewCode("NEOS")
	OMNI       = NewCode("OMNI")
	PASC       = NewCode("PASC")
	PPC        = NewCode("PPC")
	DSH        = NewCode("DSH")
	GML        = NewCode("GML")
	GSY        = NewCode("GSY")
	POT        = NewCode("POT")
	XPM        = NewCode("XPM")
	AMP        = NewCode("AMP")
	VRC        = NewCode("VRC")
	VTC        = NewCode("VTC")
	ZERO07     = NewCode("007")
	BIT16      = NewCode("BIT16")
	TWO015     = NewCode("2015")
	TWO56      = NewCode("256")
	TWOBACCO   = NewCode("2BACCO")
	TWOGIVE    = NewCode("2GIVE")
	THIRTY2BIT = NewCode("32BIT")
	THREE65    = NewCode("365")
	FOUR04     = NewCode("404")
	SEVEN00    = NewCode("700")
	EIGHTBIT   = NewCode("8BIT")
	ACLR       = NewCode("ACLR")
	ACES       = NewCode("ACES")
	ACPR       = NewCode("ACPR")
	ACID       = NewCode("ACID")
	ACOIN      = NewCode("ACOIN")
	ACRN       = NewCode("ACRN")
	ADAM       = NewCode("ADAM")
	ADT        = NewCode("ADT")
	AIB        = NewCode("AIB")
	ADZ        = NewCode("ADZ")
	AECC       = NewCode("AECC")
	AM         = NewCode("AM")
	AGRI       = NewCode("AGRI")
	AGT        = NewCode("AGT")
	AIR        = NewCode("AIR")
	ALEX       = NewCode("ALEX")
	AUM        = NewCode("AUM")
	ALIEN      = NewCode("ALIEN")
	ALIS       = NewCode("ALIS")
	ALL        = NewCode("ALL")
	ASAFE      = NewCode("ASAFE")
	AMBER      = NewCode("AMBER")
	AMS        = NewCode("AMS")
	ANAL       = NewCode("ANAL")
	ACP        = NewCode("ACP")
	ANI        = NewCode("ANI")
	ANTI       = NewCode("ANTI")
	ALTC       = NewCode("ALTC")
	APT        = NewCode("APT")
	ARCO       = NewCode("ARCO")
	ALC        = NewCode("ALC")
	ARB        = NewCode("ARB")
	ARCT       = NewCode("ARCT")
	ARCX       = NewCode("ARCX")
	ARGUS      = NewCode("ARGUS")
	ARH        = NewCode("ARH")
	ARM        = NewCode("ARM")
	ARNA       = NewCode("ARNA")
	ARPA       = NewCode("ARPA")
	ARTA       = NewCode("ARTA")
	ABY        = NewCode("ABY")
	ARTC       = NewCode("ARTC")
	AL         = NewCode("AL")
	ASN        = NewCode("ASN")
	ADCN       = NewCode("ADCN")
	ATB        = NewCode("ATB")
	ATM        = NewCode("ATM")
	ATMCHA     = NewCode("ATMCHA")
	ATOM       = NewCode("ATOM")
	ADC        = NewCode("ADC")
	ARE        = NewCode("ARE")
	AUR        = NewCode("AUR")
	AV         = NewCode("AV")
	AXIOM      = NewCode("AXIOM")
	B2B        = NewCode("B2B")
	B2         = NewCode("B2")
	B3         = NewCode("B3")
	BAB        = NewCode("BAB")
	BAN        = NewCode("BAN")
	BamitCoin  = NewCode("BamitCoin")
	NANAS      = NewCode("NANAS")
	BBCC       = NewCode("BBCC")
	BTA        = NewCode("BTA")
	BSTK       = NewCode("BSTK")
	BATL       = NewCode("BATL")
	BBH        = NewCode("BBH")
	BITB       = NewCode("BITB")
	BRDD       = NewCode("BRDD")
	XBTS       = NewCode("XBTS")
	BVC        = NewCode("BVC")
	CHATX      = NewCode("CHATX")
	BEEP       = NewCode("BEEP")
	BEEZ       = NewCode("BEEZ")
	BENJI      = NewCode("BENJI")
	BERN       = NewCode("BERN")
	PROFIT     = NewCode("PROFIT")
	BEST       = NewCode("BEST")
	BGF        = NewCode("BGF")
	BIGUP      = NewCode("BIGUP")
	BLRY       = NewCode("BLRY")
	BILL       = NewCode("BILL")
	BIOB       = NewCode("BIOB")
	BIO        = NewCode("BIO")
	BIOS       = NewCode("BIOS")
	BPTN       = NewCode("BPTN")
	BTCA       = NewCode("BTCA")
	BA         = NewCode("BA")
	BAC        = NewCode("BAC")
	BBT        = NewCode("BBT")
	BOSS       = NewCode("BOSS")
	BRONZ      = NewCode("BRONZ")
	CAT        = NewCode("CAT")
	BTD        = NewCode("BTD")
	XBTC21     = NewCode("XBTC21")
	BCA        = NewCode("BCA")
	BCP        = NewCode("BCP")
	BTDOLL     = NewCode("BTDOLL")
	LIZA       = NewCode("LIZA")
	BTCRED     = NewCode("BTCRED")
	BTCS       = NewCode("BTCS")
	BTU        = NewCode("BTU")
	BUM        = NewCode("BUM")
	LITE       = NewCode("LITE")
	BCM        = NewCode("BCM")
	BCS        = NewCode("BCS")
	BTCU       = NewCode("BTCU")
	BM         = NewCode("BM")
	BTCRY      = NewCode("BTCRY")
	BTCR       = NewCode("BTCR")
	HIRE       = NewCode("HIRE")
	STU        = NewCode("STU")
	BITOK      = NewCode("BITOK")
	BITON      = NewCode("BITON")
	BPC        = NewCode("BPC")
	BPOK       = NewCode("BPOK")
	BTP        = NewCode("BTP")
	BITCNY     = NewCode("bitCNY")
	RNTB       = NewCode("RNTB")
	BSH        = NewCode("BSH")
	XBS        = NewCode("XBS")
	BITS       = NewCode("BITS")
	BST        = NewCode("BST")
	BXT        = NewCode("BXT")
	VEG        = NewCode("VEG")
	VOLT       = NewCode("VOLT")
	BTV        = NewCode("BTV")
	BITZ       = NewCode("BITZ")
	BTZ        = NewCode("BTZ")
	BHC        = NewCode("BHC")
	BDC        = NewCode("BDC")
	JACK       = NewCode("JACK")
	BS         = NewCode("BS")
	BSTAR      = NewCode("BSTAR")
	BLAZR      = NewCode("BLAZR")
	BOD        = NewCode("BOD")
	BLUE       = NewCode("BLUE")
	BLU        = NewCode("BLU")
	BLUS       = NewCode("BLUS")
	BMT        = NewCode("BMT")
	BOLI       = NewCode("BOLI")
	BOMB       = NewCode("BOMB")
	BON        = NewCode("BON")
	BOOM       = NewCode("BOOM")
	BOSON      = NewCode("BOSON")
	BSC        = NewCode("BSC")
	BRH        = NewCode("BRH")
	BRAIN      = NewCode("BRAIN")
	BRE        = NewCode("BRE")
	BTCM       = NewCode("BTCM")
	BTCO       = NewCode("BTCO")
	TALK       = NewCode("TALK")
	BUB        = NewCode("BUB")
	BUY        = NewCode("BUY")
	BUZZ       = NewCode("BUZZ")
	BTH        = NewCode("BTH")
	C0C0       = NewCode("C0C0")
	CAB        = NewCode("CAB")
	CF         = NewCode("CF")
	CLO        = NewCode("CLO")
	CAM        = NewCode("CAM")
	CD         = NewCode("CD")
	CANN       = NewCode("CANN")
	CNNC       = NewCode("CNNC")
	CPC        = NewCode("CPC")
	CST        = NewCode("CST")
	CAPT       = NewCode("CAPT")
	CARBON     = NewCode("CARBON")
	CME        = NewCode("CME")
	CTK        = NewCode("CTK")
	CBD        = NewCode("CBD")
	CCC        = NewCode("CCC")
	CNT        = NewCode("CNT")
	XCE        = NewCode("XCE")
	CHRG       = NewCode("CHRG")
	CHEMX      = NewCode("CHEMX")
	CHESS      = NewCode("CHESS")
	CKS        = NewCode("CKS")
	CHILL      = NewCode("CHILL")
	CHIP       = NewCode("CHIP")
	CHOOF      = NewCode("CHOOF")
	CRX        = NewCode("CRX")
	CIN        = NewCode("CIN")
	POLL       = NewCode("POLL")
	CLICK      = NewCode("CLICK")
	CLINT      = NewCode("CLINT")
	CLUB       = NewCode("CLUB")
	CLUD       = NewCode("CLUD")
	COX        = NewCode("COX")
	COXST      = NewCode("COXST")
	CFC        = NewCode("CFC")
	CTIC2      = NewCode("CTIC2")
	COIN       = NewCode("COIN")
	BTTF       = NewCode("BTTF")
	C2         = NewCode("C2")
	CAID       = NewCode("CAID")
	CL         = NewCode("CL")
	CTIC       = NewCode("CTIC")
	CXT        = NewCode("CXT")
	CHP        = NewCode("CHP")
	CV2        = NewCode("CV2")
	COC        = NewCode("COC")
	COMP       = NewCode("COMP")
	CMS        = NewCode("CMS")
	CONX       = NewCode("CONX")
	CCX        = NewCode("CCX")
	CLR        = NewCode("CLR")
	CORAL      = NewCode("CORAL")
	CORG       = NewCode("CORG")
	CSMIC      = NewCode("CSMIC")
	CMC        = NewCode("CMC")
	COV        = NewCode("COV")
	COVX       = NewCode("COVX")
	CRAB       = NewCode("CRAB")
	CRAFT      = NewCode("CRAFT")
	CRNK       = NewCode("CRNK")
	CRAVE      = NewCode("CRAVE")
	CRM        = NewCode("CRM")
	XCRE       = NewCode("XCRE")
	CREDIT     = NewCode("CREDIT")
	CREVA      = NewCode("CREVA")
	CRIME      = NewCode("CRIME")
	CROC       = NewCode("CROC")
	CRC        = NewCode("CRC")
	CRW        = NewCode("CRW")
	CRY        = NewCode("CRY")
	CBX        = NewCode("CBX")
	TKTX       = NewCode("TKTX")
	CB         = NewCode("CB")
	CIRC       = NewCode("CIRC")
	CCB        = NewCode("CCB")
	CDO        = NewCode("CDO")
	CG         = NewCode("CG")
	CJ         = NewCode("CJ")
	CJC        = NewCode("CJC")
	CYT        = NewCode("CYT")
	CRPS       = NewCode("CRPS")
	PING       = NewCode("PING")
	CWXT       = NewCode("CWXT")
	CCT        = NewCode("CCT")
	CTL        = NewCode("CTL")
	CURVES     = NewCode("CURVES")
	CC         = NewCode("CC")
	CYC        = NewCode("CYC")
	CYG        = NewCode("CYG")
	CYP        = NewCode("CYP")
	FUNK       = NewCode("FUNK")
	CZECO      = NewCode("CZECO")
	DALC       = NewCode("DALC")
	DLISK      = NewCode("DLISK")
	MOOND      = NewCode("MOOND")
	DB         = NewCode("DB")
	DCC        = NewCode("DCC")
	DCYP       = NewCode("DCYP")
	DETH       = NewCode("DETH")
	DKC        = NewCode("DKC")
	DISK       = NewCode("DISK")
	DRKT       = NewCode("DRKT")
	DTT        = NewCode("DTT")
	DASHS      = NewCode("DASHS")
	DBTC       = NewCode("DBTC")
	DCT        = NewCode("DCT")
	DBET       = NewCode("DBET")
	DEC        = NewCode("DEC")
	DECR       = NewCode("DECR")
	DEA        = NewCode("DEA")
	DPAY       = NewCode("DPAY")
	DCRE       = NewCode("DCRE")
	DC         = NewCode("DC")
	DES        = NewCode("DES")
	DEM        = NewCode("DEM")
	DXC        = NewCode("DXC")
	DCK        = NewCode("DCK")
	CUBE       = NewCode("CUBE")
	DGMS       = NewCode("DGMS")
	DBG        = NewCode("DBG")
	DGCS       = NewCode("DGCS")
	DBLK       = NewCode("DBLK")
	DIME       = NewCode("DIME")
	DIRT       = NewCode("DIRT")
	DVD        = NewCode("DVD")
	DMT        = NewCode("DMT")
	NOTE       = NewCode("NOTE")
	DGORE      = NewCode("DGORE")
	DLC        = NewCode("DLC")
	DRT        = NewCode("DRT")
	DOTA       = NewCode("DOTA")
	DOX        = NewCode("DOX")
	DRA        = NewCode("DRA")
	DFT        = NewCode("DFT")
	XDB        = NewCode("XDB")
	DRM        = NewCode("DRM")
	DRZ        = NewCode("DRZ")
	DRACO      = NewCode("DRACO")
	DBIC       = NewCode("DBIC")
	DUB        = NewCode("DUB")
	GUM        = NewCode("GUM")
	DUR        = NewCode("DUR")
	DUST       = NewCode("DUST")
	DUX        = NewCode("DUX")
	DXO        = NewCode("DXO")
	ECN        = NewCode("ECN")
	EDR2       = NewCode("EDR2")
	EA         = NewCode("EA")
	EAGS       = NewCode("EAGS")
	EMT        = NewCode("EMT")
	EBONUS     = NewCode("EBONUS")
	ECCHI      = NewCode("ECCHI")
	EKO        = NewCode("EKO")
	ECLI       = NewCode("ECLI")
	ECOB       = NewCode("ECOB")
	ECO        = NewCode("ECO")
	EDIT       = NewCode("EDIT")
	EDRC       = NewCode("EDRC")
	EDC        = NewCode("EDC")
	EGAME      = NewCode("EGAME")
	EGG        = NewCode("EGG")
	EGO        = NewCode("EGO")
	ELC        = NewCode("ELC")
	ELCO       = NewCode("ELCO")
	ECA        = NewCode("ECA")
	EPC        = NewCode("EPC")
	ELE        = NewCode("ELE")
	ONE337     = NewCode("1337")
	EMB        = NewCode("EMB")
	EMC        = NewCode("EMC")
	EPY        = NewCode("EPY")
	EMPC       = NewCode("EMPC")
	EMP        = NewCode("EMP")
	ENE        = NewCode("ENE")
	EET        = NewCode("EET")
	XNG        = NewCode("XNG")
	EGMA       = NewCode("EGMA")
	ENTER      = NewCode("ENTER")
	ETRUST     = NewCode("ETRUST")
	EQL        = NewCode("EQL")
	EQM        = NewCode("EQM")
	EQT        = NewCode("EQT")
	ERR        = NewCode("ERR")
	ESC        = NewCode("ESC")
	ESP        = NewCode("ESP")
	ENT        = NewCode("ENT")
	ETCO       = NewCode("ETCO")
	DOGETH     = NewCode("DOGETH")
	ECASH      = NewCode("ECASH")
	ELITE      = NewCode("ELITE")
	ETHS       = NewCode("ETHS")
	ETL        = NewCode("ETL")
	ETZ        = NewCode("ETZ")
	EUC        = NewCode("EUC")
	EURC       = NewCode("EURC")
	EUROPE     = NewCode("EUROPE")
	EVA        = NewCode("EVA")
	EGC        = NewCode("EGC")
	EOC        = NewCode("EOC")
	EVIL       = NewCode("EVIL")
	EVO        = NewCode("EVO")
	EXB        = NewCode("EXB")
	EXIT       = NewCode("EXIT")
	XT         = NewCode("XT")
	F16        = NewCode("F16")
	FADE       = NewCode("FADE")
	FAZZ       = NewCode("FAZZ")
	FX         = NewCode("FX")
	FIDEL      = NewCode("FIDEL")
	FIDGT      = NewCode("FIDGT")
	FIND       = NewCode("FIND")
	FPC        = NewCode("FPC")
	FIRE       = NewCode("FIRE")
	FFC        = NewCode("FFC")
	FRST       = NewCode("FRST")
	FIST       = NewCode("FIST")
	FIT        = NewCode("FIT")
	FLX        = NewCode("FLX")
	FLVR       = NewCode("FLVR")
	FLY        = NewCode("FLY")
	FONZ       = NewCode("FONZ")
	XFCX       = NewCode("XFCX")
	FOREX      = NewCode("FOREX")
	FRN        = NewCode("FRN")
	FRK        = NewCode("FRK")
	FRWC       = NewCode("FRWC")
	FGZ        = NewCode("FGZ")
	FRE        = NewCode("FRE")
	FRDC       = NewCode("FRDC")
	FJC        = NewCode("FJC")
	FURY       = NewCode("FURY")
	FSN        = NewCode("FSN")
	FCASH      = NewCode("FCASH")
	FTO        = NewCode("FTO")
	FUZZ       = NewCode("FUZZ")
	GAKH       = NewCode("GAKH")
	GBT        = NewCode("GBT")
	UNITS      = NewCode("UNITS")
	FOUR20G    = NewCode("420G")
	GENIUS     = NewCode("GENIUS")
	GEN        = NewCode("GEN")
	GEO        = NewCode("GEO")
	GER        = NewCode("GER")
	GSR        = NewCode("GSR")
	SPKTR      = NewCode("SPKTR")
	GIFT       = NewCode("GIFT")
	WTT        = NewCode("WTT")
	GHS        = NewCode("GHS")
	GIG        = NewCode("GIG")
	GOT        = NewCode("GOT")
	XGTC       = NewCode("XGTC")
	GIZ        = NewCode("GIZ")
	GLO        = NewCode("GLO")
	GCR        = NewCode("GCR")
	BSTY       = NewCode("BSTY")
	GLC        = NewCode("GLC")
	GSX        = NewCode("GSX")
	GOAT       = NewCode("GOAT")
	GB         = NewCode("GB")
	GFL        = NewCode("GFL")
	MNTP       = NewCode("MNTP")
	GP         = NewCode("GP")
	GLUCK      = NewCode("GLUCK")
	GOON       = NewCode("GOON")
	GTFO       = NewCode("GTFO")
	GOTX       = NewCode("GOTX")
	GPU        = NewCode("GPU")
	GRF        = NewCode("GRF")
	GRAM       = NewCode("GRAM")
	GRAV       = NewCode("GRAV")
	GBIT       = NewCode("GBIT")
	GREED      = NewCode("GREED")
	GE         = NewCode("GE")
	GREENF     = NewCode("GREENF")
	GRE        = NewCode("GRE")
	GREXIT     = NewCode("GREXIT")
	GMCX       = NewCode("GMCX")
	GROW       = NewCode("GROW")
	GSM        = NewCode("GSM")
	GT         = NewCode("GT")
	NLG        = NewCode("NLG")
	HKN        = NewCode("HKN")
	HAC        = NewCode("HAC")
	HALLO      = NewCode("HALLO")
	HAMS       = NewCode("HAMS")
	HPC        = NewCode("HPC")
	HAWK       = NewCode("HAWK")
	HAZE       = NewCode("HAZE")
	HZT        = NewCode("HZT")
	HDG        = NewCode("HDG")
	HEDG       = NewCode("HEDG")
	HEEL       = NewCode("HEEL")
	HMP        = NewCode("HMP")
	PLAY       = NewCode("PLAY")
	HXX        = NewCode("HXX")
	XHI        = NewCode("XHI")
	HVCO       = NewCode("HVCO")
	HTC        = NewCode("HTC")
	MINH       = NewCode("MINH")
	HODL       = NewCode("HODL")
	HON        = NewCode("HON")
	HOPE       = NewCode("HOPE")
	HQX        = NewCode("HQX")
	HSP        = NewCode("HSP")
	HTML5      = NewCode("HTML5")
	HYPERX     = NewCode("HYPERX")
	HPS        = NewCode("HPS")
	IOC        = NewCode("IOC")
	IBANK      = NewCode("IBANK")
	IBITS      = NewCode("IBITS")
	ICASH      = NewCode("ICASH")
	ICOB       = NewCode("ICOB")
	ICON       = NewCode("ICON")
	IETH       = NewCode("IETH")
	ILM        = NewCode("ILM")
	IMPS       = NewCode("IMPS")
	NKA        = NewCode("NKA")
	INCP       = NewCode("INCP")
	IN         = NewCode("IN")
	INC        = NewCode("INC")
	IMS        = NewCode("IMS")
	IFLT       = NewCode("IFLT")
	INFX       = NewCode("INFX")
	INGT       = NewCode("INGT")
	INPAY      = NewCode("INPAY")
	INSANE     = NewCode("INSANE")
	INXT       = NewCode("INXT")
	IFT        = NewCode("IFT")
	INV        = NewCode("INV")
	IVZ        = NewCode("IVZ")
	ILT        = NewCode("ILT")
	IONX       = NewCode("IONX")
	ISL        = NewCode("ISL")
	ITI        = NewCode("ITI")
	ING        = NewCode("ING")
	IEC        = NewCode("IEC")
	IW         = NewCode("IW")
	IXC        = NewCode("IXC")
	IXT        = NewCode("IXT")
	JPC        = NewCode("JPC")
	JANE       = NewCode("JANE")
	JWL        = NewCode("JWL")
	JIF        = NewCode("JIF")
	JOBS       = NewCode("JOBS")
	JOCKER     = NewCode("JOCKER")
	JW         = NewCode("JW")
	JOK        = NewCode("JOK")
	XJO        = NewCode("XJO")
	KGB        = NewCode("KGB")
	KARMC      = NewCode("KARMC")
	KARMA      = NewCode("KARMA")
	KASHH      = NewCode("KASHH")
	KAT        = NewCode("KAT")
	KC         = NewCode("KC")
	KIDS       = NewCode("KIDS")
	KIN        = NewCode("KIN")
	KISS       = NewCode("KISS")
	KOBO       = NewCode("KOBO")
	TP1        = NewCode("TP1")
	KRAK       = NewCode("KRAK")
	KGC        = NewCode("KGC")
	KTK        = NewCode("KTK")
	KR         = NewCode("KR")
	KUBO       = NewCode("KUBO")
	KURT       = NewCode("KURT")
	KUSH       = NewCode("KUSH")
	LANA       = NewCode("LANA")
	LTH        = NewCode("LTH")
	LAZ        = NewCode("LAZ")
	LEA        = NewCode("LEA")
	LEAF       = NewCode("LEAF")
	LENIN      = NewCode("LENIN")
	LEPEN      = NewCode("LEPEN")
	LIR        = NewCode("LIR")
	LVG        = NewCode("LVG")
	LGBTQ      = NewCode("LGBTQ")
	LHC        = NewCode("LHC")
	EXT        = NewCode("EXT")
	LBTC       = NewCode("LBTC")
	LSD        = NewCode("LSD")
	LIMX       = NewCode("LIMX")
	LTD        = NewCode("LTD")
	LINDA      = NewCode("LINDA")
	LKC        = NewCode("LKC")
	LBTCX      = NewCode("LBTCX")
	LCC        = NewCode("LCC")
	LTCU       = NewCode("LTCU")
	LTCR       = NewCode("LTCR")
	LDOGE      = NewCode("LDOGE")
	LTS        = NewCode("LTS")
	LIV        = NewCode("LIV")
	LIZI       = NewCode("LIZI")
	LOC        = NewCode("LOC")
	LOCX       = NewCode("LOCX")
	LOOK       = NewCode("LOOK")
	LOOT       = NewCode("LOOT")
	XLTCG      = NewCode("XLTCG")
	BASH       = NewCode("BASH")
	LUCKY      = NewCode("LUCKY")
	L7S        = NewCode("L7S")
	LDM        = NewCode("LDM")
	LUMI       = NewCode("LUMI")
	LUNA       = NewCode("LUNA")
	LC         = NewCode("LC")
	LUX        = NewCode("LUX")
	MCRN       = NewCode("MCRN")
	XMG        = NewCode("XMG")
	MMXIV      = NewCode("MMXIV")
	MAT        = NewCode("MAT")
	MAO        = NewCode("MAO")
	MAPC       = NewCode("MAPC")
	MRB        = NewCode("MRB")
	MXT        = NewCode("MXT")
	MARV       = NewCode("MARV")
	MARX       = NewCode("MARX")
	MCAR       = NewCode("MCAR")
	MM         = NewCode("MM")
	MVC        = NewCode("MVC")
	MAVRO      = NewCode("MAVRO")
	MAX        = NewCode("MAX")
	MAZE       = NewCode("MAZE")
	MBIT       = NewCode("MBIT")
	MCOIN      = NewCode("MCOIN")
	MPRO       = NewCode("MPRO")
	XMS        = NewCode("XMS")
	MLITE      = NewCode("MLITE")
	MLNC       = NewCode("MLNC")
	MENTAL     = NewCode("MENTAL")
	MERGEC     = NewCode("MERGEC")
	MTLMC3     = NewCode("MTLMC3")
	METAL      = NewCode("METAL")
	MUU        = NewCode("MUU")
	MILO       = NewCode("MILO")
	MND        = NewCode("MND")
	XMINE      = NewCode("XMINE")
	MNM        = NewCode("MNM")
	XNM        = NewCode("XNM")
	MIRO       = NewCode("MIRO")
	MIS        = NewCode("MIS")
	MMXVI      = NewCode("MMXVI")
	MOIN       = NewCode("MOIN")
	MOJO       = NewCode("MOJO")
	TAB        = NewCode("TAB")
	MONETA     = NewCode("MONETA")
	MUE        = NewCode("MUE")
	MONEY      = NewCode("MONEY")
	MRP        = NewCode("MRP")
	MOTO       = NewCode("MOTO")
	MULTI      = NewCode("MULTI")
	MST        = NewCode("MST")
	MVR        = NewCode("MVR")
	MYSTIC     = NewCode("MYSTIC")
	WISH       = NewCode("WISH")
	NKT        = NewCode("NKT")
	NAT        = NewCode("NAT")
	ENAU       = NewCode("ENAU")
	NEBU       = NewCode("NEBU")
	NEF        = NewCode("NEF")
	NBIT       = NewCode("NBIT")
	NETKO      = NewCode("NETKO")
	NTM        = NewCode("NTM")
	NETC       = NewCode("NETC")
	NRC        = NewCode("NRC")
	NTK        = NewCode("NTK")
	NTRN       = NewCode("NTRN")
	NEVA       = NewCode("NEVA")
	NIC        = NewCode("NIC")
	NKC        = NewCode("NKC")
	NYC        = NewCode("NYC")
	NZC        = NewCode("NZC")
	NICE       = NewCode("NICE")
	NDOGE      = NewCode("NDOGE")
	XTR        = NewCode("XTR")
	N2O        = NewCode("N2O")
	NIXON      = NewCode("NIXON")
	NOC        = NewCode("NOC")
	NODC       = NewCode("NODC")
	NODES      = NewCode("NODES")
	NODX       = NewCode("NODX")
	NLC        = NewCode("NLC")
	NLC2       = NewCode("NLC2")
	NOO        = NewCode("NOO")
	NVC        = NewCode("NVC")
	NPC        = NewCode("NPC")
	NUBIS      = NewCode("NUBIS")
	NUKE       = NewCode("NUKE")
	N7         = NewCode("N7")
	NUM        = NewCode("NUM")
	NMR        = NewCode("NMR")
	NXE        = NewCode("NXE")
	OBS        = NewCode("OBS")
	OCEAN      = NewCode("OCEAN")
	OCOW       = NewCode("OCOW")
	EIGHT88    = NewCode("888")
	OCC        = NewCode("OCC")
	OK         = NewCode("OK")
	ODNT       = NewCode("ODNT")
	FLAV       = NewCode("FLAV")
	OLIT       = NewCode("OLIT")
	OLYMP      = NewCode("OLYMP")
	OMA        = NewCode("OMA")
	OMC        = NewCode("OMC")
	ONEK       = NewCode("ONEK")
	ONX        = NewCode("ONX")
	XPO        = NewCode("XPO")
	OPAL       = NewCode("OPAL")
	OTN        = NewCode("OTN")
	OP         = NewCode("OP")
	OPES       = NewCode("OPES")
	OPTION     = NewCode("OPTION")
	ORLY       = NewCode("ORLY")
	OS76       = NewCode("OS76")
	OZC        = NewCode("OZC")
	P7C        = NewCode("P7C")
	PAC        = NewCode("PAC")
	PAK        = NewCode("PAK")
	PAL        = NewCode("PAL")
	PND        = NewCode("PND")
	PINKX      = NewCode("PINKX")
	POPPY      = NewCode("POPPY")
	DUO        = NewCode("DUO")
	PARA       = NewCode("PARA")
	PKB        = NewCode("PKB")
	GENE       = NewCode("GENE")
	PARTY      = NewCode("PARTY")
	PYN        = NewCode("PYN")
	XPY        = NewCode("XPY")
	CON        = NewCode("CON")
	PAYP       = NewCode("PAYP")
	GUESS      = NewCode("GUESS")
	PEN        = NewCode("PEN")
	PTA        = NewCode("PTA")
	PEO        = NewCode("PEO")
	PSB        = NewCode("PSB")
	XPD        = NewCode("XPD")
	PXL        = NewCode("PXL")
	PHR        = NewCode("PHR")
	PIE        = NewCode("PIE")
	PIO        = NewCode("PIO")
	PIPR       = NewCode("PIPR")
	SKULL      = NewCode("SKULL")
	PLANET     = NewCode("PLANET")
	PNC        = NewCode("PNC")
	XPTX       = NewCode("XPTX")
	PLNC       = NewCode("PLNC")
	XPS        = NewCode("XPS")
	POKE       = NewCode("POKE")
	PLBT       = NewCode("PLBT")
	POM        = NewCode("POM")
	PONZ2      = NewCode("PONZ2")
	PONZI      = NewCode("PONZI")
	XSP        = NewCode("XSP")
	XPC        = NewCode("XPC")
	PEX        = NewCode("PEX")
	TRON       = NewCode("TRON")
	POST       = NewCode("POST")
	POSW       = NewCode("POSW")
	PWR        = NewCode("PWR")
	POWER      = NewCode("POWER")
	PRE        = NewCode("PRE")
	PRS        = NewCode("PRS")
	PXI        = NewCode("PXI")
	PEXT       = NewCode("PEXT")
	PRIMU      = NewCode("PRIMU")
	PRX        = NewCode("PRX")
	PRM        = NewCode("PRM")
	PRIX       = NewCode("PRIX")
	XPRO       = NewCode("XPRO")
	PCM        = NewCode("PCM")
	PROC       = NewCode("PROC")
	NANOX      = NewCode("NANOX")
	VRP        = NewCode("VRP")
	PTY        = NewCode("PTY")
	PSI        = NewCode("PSI")
	PSY        = NewCode("PSY")
	PULSE      = NewCode("PULSE")
	PUPA       = NewCode("PUPA")
	PURE       = NewCode("PURE")
	VIDZ       = NewCode("VIDZ")
	PUTIN      = NewCode("PUTIN")
	PX         = NewCode("PX")
	QTM        = NewCode("QTM")
	QTZ        = NewCode("QTZ")
	QBC        = NewCode("QBC")
	XQN        = NewCode("XQN")
	RBBT       = NewCode("RBBT")
	RAC        = NewCode("RAC")
	RADI       = NewCode("RADI")
	RAD        = NewCode("RAD")
	RAI        = NewCode("RAI")
	XRA        = NewCode("XRA")
	RATIO      = NewCode("RATIO")
	REA        = NewCode("REA")
	RCX        = NewCode("RCX")
	REE        = NewCode("REE")
	REC        = NewCode("REC")
	RMS        = NewCode("RMS")
	RBIT       = NewCode("RBIT")
	RNC        = NewCode("RNC")
	REV        = NewCode("REV")
	RH         = NewCode("RH")
	XRL        = NewCode("XRL")
	RICE       = NewCode("RICE")
	RICHX      = NewCode("RICHX")
	RID        = NewCode("RID")
	RIDE       = NewCode("RIDE")
	RBT        = NewCode("RBT")
	RING       = NewCode("RING")
	RIO        = NewCode("RIO")
	RISE       = NewCode("RISE")
	ROCKET     = NewCode("ROCKET")
	RPC        = NewCode("RPC")
	ROS        = NewCode("ROS")
	ROYAL      = NewCode("ROYAL")
	RSGP       = NewCode("RSGP")
	RBIES      = NewCode("RBIES")
	RUBIT      = NewCode("RUBIT")
	RBY        = NewCode("RBY")
	RUC        = NewCode("RUC")
	RUPX       = NewCode("RUPX")
	RUP        = NewCode("RUP")
	RUST       = NewCode("RUST")
	SFE        = NewCode("SFE")
	SLS        = NewCode("SLS")
	SMSR       = NewCode("SMSR")
	RONIN      = NewCode("RONIN")
	STV        = NewCode("STV")
	HIFUN      = NewCode("HIFUN")
	MAD        = NewCode("MAD")
	SANDG      = NewCode("SANDG")
	STO        = NewCode("STO")
	SCAN       = NewCode("SCAN")
	SCITW      = NewCode("SCITW")
	SCRPT      = NewCode("SCRPT")
	SCRT       = NewCode("SCRT")
	SED        = NewCode("SED")
	SEEDS      = NewCode("SEEDS")
	B2X        = NewCode("B2X")
	SEL        = NewCode("SEL")
	SLFI       = NewCode("SLFI")
	SMBR       = NewCode("SMBR")
	SEN        = NewCode("SEN")
	SENT       = NewCode("SENT")
	SRNT       = NewCode("SRNT")
	SEV        = NewCode("SEV")
	SP         = NewCode("SP")
	SXC        = NewCode("SXC")
	GELD       = NewCode("GELD")
	SHDW       = NewCode("SHDW")
	SDC        = NewCode("SDC")
	SAK        = NewCode("SAK")
	SHRP       = NewCode("SHRP")
	SHELL      = NewCode("SHELL")
	SH         = NewCode("SH")
	SHORTY     = NewCode("SHORTY")
	SHREK      = NewCode("SHREK")
	SHRM       = NewCode("SHRM")
	SIB        = NewCode("SIB")
	SIGT       = NewCode("SIGT")
	SLCO       = NewCode("SLCO")
	SIGU       = NewCode("SIGU")
	SIX        = NewCode("SIX")
	SJW        = NewCode("SJW")
	SKB        = NewCode("SKB")
	SW         = NewCode("SW")
	SLEEP      = NewCode("SLEEP")
	SLING      = NewCode("SLING")
	SMART      = NewCode("SMART")
	SMC        = NewCode("SMC")
	SMF        = NewCode("SMF")
	SOCC       = NewCode("SOCC")
	SCL        = NewCode("SCL")
	SDAO       = NewCode("SDAO")
	SOLAR      = NewCode("SOLAR")
	SOLO       = NewCode("SOLO")
	SCT        = NewCode("SCT")
	SONG       = NewCode("SONG")
	ALTCOM     = NewCode("ALTCOM")
	SPHTX      = NewCode("SPHTX")
	SPC        = NewCode("SPC")
	SPACE      = NewCode("SPACE")
	SBT        = NewCode("SBT")
	SPEC       = NewCode("SPEC")
	SPX        = NewCode("SPX")
	SCS        = NewCode("SCS")
	SPORT      = NewCode("SPORT")
	SPT        = NewCode("SPT")
	SPR        = NewCode("SPR")
	SPEX       = NewCode("SPEX")
	SQL        = NewCode("SQL")
	SBIT       = NewCode("SBIT")
	STHR       = NewCode("STHR")
	STALIN     = NewCode("STALIN")
	STAR       = NewCode("STAR")
	STA        = NewCode("STA")
	START      = NewCode("START")
	STP        = NewCode("STP")
	PNK        = NewCode("PNK")
	STEPS      = NewCode("STEPS")
	STK        = NewCode("STK")
	STONK      = NewCode("STONK")
	STS        = NewCode("STS")
	STRP       = NewCode("STRP")
	STY        = NewCode("STY")
	XMT        = NewCode("XMT")
	SSTC       = NewCode("SSTC")
	SUPER      = NewCode("SUPER")
	SRND       = NewCode("SRND")
	STRB       = NewCode("STRB")
	M1         = NewCode("M1")
	SPM        = NewCode("SPM")
	BUCKS      = NewCode("BUCKS")
	TOKEN      = NewCode("TOKEN")
	SWT        = NewCode("SWT")
	SWEET      = NewCode("SWEET")
	SWING      = NewCode("SWING")
	CHSB       = NewCode("CHSB")
	SIC        = NewCode("SIC")
	SDP        = NewCode("SDP")
	XSY        = NewCode("XSY")
	SYNX       = NewCode("SYNX")
	SNRG       = NewCode("SNRG")
	TAG        = NewCode("TAG")
	TAGR       = NewCode("TAGR")
	TAJ        = NewCode("TAJ")
	TAK        = NewCode("TAK")
	TAKE       = NewCode("TAKE")
	TAM        = NewCode("TAM")
	XTO        = NewCode("XTO")
	TAP        = NewCode("TAP")
	TLE        = NewCode("TLE")
	TSE        = NewCode("TSE")
	TLEX       = NewCode("TLEX")
	TAXI       = NewCode("TAXI")
	TCN        = NewCode("TCN")
	TDFB       = NewCode("TDFB")
	TEAM       = NewCode("TEAM")
	TECH       = NewCode("TECH")
	TEC        = NewCode("TEC")
	TEK        = NewCode("TEK")
	TB         = NewCode("TB")
	TLX        = NewCode("TLX")
	TELL       = NewCode("TELL")
	TENNET     = NewCode("TENNET")
	TES        = NewCode("TES")
	TGS        = NewCode("TGS")
	XVE        = NewCode("XVE")
	TCR        = NewCode("TCR")
	GCC        = NewCode("GCC")
	MAY        = NewCode("MAY")
	THOM       = NewCode("THOM")
	TIA        = NewCode("TIA")
	TIDE       = NewCode("TIDE")
	TIE        = NewCode("TIE")
	TIT        = NewCode("TIT")
	TTC        = NewCode("TTC")
	TODAY      = NewCode("TODAY")
	TBX        = NewCode("TBX")
	TDS        = NewCode("TDS")
	TLOSH      = NewCode("TLOSH")
	TOKC       = NewCode("TOKC")
	TMRW       = NewCode("TMRW")
	TOOL       = NewCode("TOOL")
	TCX        = NewCode("TCX")
	TOT        = NewCode("TOT")
	TX         = NewCode("TX")
	TRANSF     = NewCode("TRANSF")
	TRAP       = NewCode("TRAP")
	TBCX       = NewCode("TBCX")
	TRICK      = NewCode("TRICK")
	TPG        = NewCode("TPG")
	TFL        = NewCode("TFL")
	TRUMP      = NewCode("TRUMP")
	TNG        = NewCode("TNG")
	TUR        = NewCode("TUR")
	TWERK      = NewCode("TWERK")
	TWIST      = NewCode("TWIST")
	TWO        = NewCode("TWO")
	UCASH      = NewCode("UCASH")
	UAE        = NewCode("UAE")
	XBU        = NewCode("XBU")
	UBQ        = NewCode("UBQ")
	U          = NewCode("U")
	UDOWN      = NewCode("UDOWN")
	GAIN       = NewCode("GAIN")
	USC        = NewCode("USC")
	UMC        = NewCode("UMC")
	UNF        = NewCode("UNF")
	UNIFY      = NewCode("UNIFY")
	USDE       = NewCode("USDE")
	UBTC       = NewCode("UBTC")
	UIS        = NewCode("UIS")
	UNIT       = NewCode("UNIT")
	UNI        = NewCode("UNI")
	UXC        = NewCode("UXC")
	URC        = NewCode("URC")
	XUP        = NewCode("XUP")
	UFR        = NewCode("UFR")
	URO        = NewCode("URO")
	UTLE       = NewCode("UTLE")
	VAL        = NewCode("VAL")
	VPRC       = NewCode("VPRC")
	VAPOR      = NewCode("VAPOR")
	VCOIN      = NewCode("VCOIN")
	VEC        = NewCode("VEC")
	VEC2       = NewCode("VEC2")
	VLT        = NewCode("VLT")
	VENE       = NewCode("VENE")
	VNTX       = NewCode("VNTX")
	VTN        = NewCode("VTN")
	CRED       = NewCode("CRED")
	VERS       = NewCode("VERS")
	VTX        = NewCode("VTX")
	VTY        = NewCode("VTY")
	VIP        = NewCode("VIP")
	VISIO      = NewCode("VISIO")
	VK         = NewCode("VK")
	VOL        = NewCode("VOL")
	VOYA       = NewCode("VOYA")
	VPN        = NewCode("VPN")
	XVS        = NewCode("XVS")
	VTL        = NewCode("VTL")
	VULC       = NewCode("VULC")
	VVI        = NewCode("VVI")
	WGR        = NewCode("WGR")
	WAM        = NewCode("WAM")
	WARP       = NewCode("WARP")
	WASH       = NewCode("WASH")
	WGO        = NewCode("WGO")
	WAY        = NewCode("WAY")
	WCASH      = NewCode("WCASH")
	WEALTH     = NewCode("WEALTH")
	WEEK       = NewCode("WEEK")
	WHO        = NewCode("WHO")
	WIC        = NewCode("WIC")
	WBB        = NewCode("WBB")
	WINE       = NewCode("WINE")
	WINK       = NewCode("WINK")
	WISC       = NewCode("WISC")
	WITCH      = NewCode("WITCH")
	WMC        = NewCode("WMC")
	WOMEN      = NewCode("WOMEN")
	WOK        = NewCode("WOK")
	WRT        = NewCode("WRT")
	XCO        = NewCode("XCO")
	X2         = NewCode("X2")
	XNX        = NewCode("XNX")
	XAU        = NewCode("XAU")
	XAV        = NewCode("XAV")
	XDE2       = NewCode("XDE2")
	XDE        = NewCode("XDE")
	XIOS       = NewCode("XIOS")
	XOC        = NewCode("XOC")
	XSSX       = NewCode("XSSX")
	XBY        = NewCode("XBY")
	YAC        = NewCode("YAC")
	YMC        = NewCode("YMC")
	YAY        = NewCode("YAY")
	YBC        = NewCode("YBC")
	YES        = NewCode("YES")
	YOB2X      = NewCode("YOB2X")
	YOVI       = NewCode("YOVI")
	ZYD        = NewCode("ZYD")
	ZECD       = NewCode("ZECD")
	ZEIT       = NewCode("ZEIT")
	ZENI       = NewCode("ZENI")
	ZET2       = NewCode("ZET2")
	ZET        = NewCode("ZET")
	ZMC        = NewCode("ZMC")
	ZIRK       = NewCode("ZIRK")
	ZLQ        = NewCode("ZLQ")
	ZNE        = NewCode("ZNE")
	ZONTO      = NewCode("ZONTO")
	ZOOM       = NewCode("ZOOM")
	ZRC        = NewCode("ZRC")
	ZUR        = NewCode("ZUR")
	ZB         = NewCode("ZB")
	QC         = NewCode("QC")
	HLC        = NewCode("HLC")
	SAFE       = NewCode("SAFE")
	BTN        = NewCode("BTN")
	CDC        = NewCode("CDC")
	DDM        = NewCode("DDM")
	HOTC       = NewCode("HOTC")
	BDS        = NewCode("BDS")
	AAA        = NewCode("AAA")
	XWC        = NewCode("XWC")
	PDX        = NewCode("PDX")
	SLT        = NewCode("SLT")
	HPY        = NewCode("HPY")
	XXBT       = NewCode("XXBT") // BTC, but XXBT instead
	XDG        = NewCode("XDG")  // DOGE
	HKD        = NewCode("HKD")  // Hong Kong Dollar
	AUD        = NewCode("AUD")  // Australian Dollar
	USD        = NewCode("USD")  // United States Dollar
	ZUSD       = NewCode("ZUSD") // United States Dollar, but with a Z in front of it
	EUR        = NewCode("EUR")  // Euro
	ZEUR       = NewCode("ZEUR") // Euro, but with a Z in front of it
	CAD        = NewCode("CAD")  // Canadaian Dollar
	ZCAD       = NewCode("ZCAD") // Canadaian Dollar, but with a Z in front of it
	SGD        = NewCode("SGD")  // Singapore Dollar
	RUB        = NewCode("RUB")  // RUssian ruBle
	RUR        = NewCode("RUR")  // RUssian Ruble
	PLN        = NewCode("PLN")  // Polish złoty
	TRY        = NewCode("TRY")  // Turkish lira
	UAH        = NewCode("UAH")  // Ukrainian hryvnia
	JPY        = NewCode("JPY")  // Japanese yen
	ZJPY       = NewCode("ZJPY") // Japanese yen, but with a Z in front of it
	LCH        = NewCode("LCH")
	MYR        = NewCode("MYR")
	AFN        = NewCode("AFN")
	ARS        = NewCode("ARS")
	AWG        = NewCode("AWG")
	AZN        = NewCode("AZN")
	BSD        = NewCode("BSD")
	BBD        = NewCode("BBD")
	BYN        = NewCode("BYN")
	BZD        = NewCode("BZD")
	BMD        = NewCode("BMD")
	BOB        = NewCode("BOB")
	BAM        = NewCode("BAM")
	BWP        = NewCode("BWP")
	BGN        = NewCode("BGN")
	BRL        = NewCode("BRL")
	BND        = NewCode("BND")
	KHR        = NewCode("KHR")
	KYD        = NewCode("KYD")
	CLP        = NewCode("CLP")
	CNY        = NewCode("CNY")
	COP        = NewCode("COP")
	HRK        = NewCode("HRK")
	CUP        = NewCode("CUP")
	CZK        = NewCode("CZK")
	DKK        = NewCode("DKK")
	DOP        = NewCode("DOP")
	XCD        = NewCode("XCD")
	EGP        = NewCode("EGP")
	SVC        = NewCode("SVC")
	FKP        = NewCode("FKP")
	FJD        = NewCode("FJD")
	GIP        = NewCode("GIP")
	GTQ        = NewCode("GTQ")
	GGP        = NewCode("GGP")
	GYD        = NewCode("GYD")
	HNL        = NewCode("HNL")
	HUF        = NewCode("HUF")
	ISK        = NewCode("ISK")
	INR        = NewCode("INR")
	IDR        = NewCode("IDR")
	IRR        = NewCode("IRR")
	IMP        = NewCode("IMP")
	ILS        = NewCode("ILS")
	JMD        = NewCode("JMD")
	JEP        = NewCode("JEP")
	KZT        = NewCode("KZT")
	KPW        = NewCode("KPW")
	KGS        = NewCode("KGS")
	LAK        = NewCode("LAK")
	LBP        = NewCode("LBP")
	LRD        = NewCode("LRD")
	MKD        = NewCode("MKD")
	MUR        = NewCode("MUR")
	MXN        = NewCode("MXN")
	MNT        = NewCode("MNT")
	MZN        = NewCode("MZN")
	NAD        = NewCode("NAD")
	NPR        = NewCode("NPR")
	ANG        = NewCode("ANG")
	NZD        = NewCode("NZD")
	NIO        = NewCode("NIO")
	NGN        = NewCode("NGN")
	NOK        = NewCode("NOK")
	OMR        = NewCode("OMR")
	PKR        = NewCode("PKR")
	PAB        = NewCode("PAB")
	PYG        = NewCode("PYG")
	PHP        = NewCode("PHP")
	QAR        = NewCode("QAR")
	RON        = NewCode("RON")
	SHP        = NewCode("SHP")
	SAR        = NewCode("SAR")
	RSD        = NewCode("RSD")
	SCR        = NewCode("SCR")
	SOS        = NewCode("SOS")
	ZAR        = NewCode("ZAR")
	LKR        = NewCode("LKR")
	SEK        = NewCode("SEK")
	CHF        = NewCode("CHF")
	SRD        = NewCode("SRD")
	SYP        = NewCode("SYP")
	TWD        = NewCode("TWD")
	THB        = NewCode("THB")
	TTD        = NewCode("TTD")
	TVD        = NewCode("TVD")
	GBP        = NewCode("GBP")
	UYU        = NewCode("UYU")
	UZS        = NewCode("UZS")
	VEF        = NewCode("VEF")
	VND        = NewCode("VND")
	YER        = NewCode("YER")
	ZWD        = NewCode("ZWD")
	XETH       = NewCode("XETH")
	FX_BTC     = NewCode("FX_BTC") // nolint: stylecheck, golint
)
