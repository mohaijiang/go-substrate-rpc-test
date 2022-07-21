package test

import (
	"fmt"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestApi(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI("ws://183.66.65.207:49944")
	if err != nil {
		panic(err)
	}

	chain, err := api.RPC.System.Chain()
	if err != nil {
		panic(err)
	}
	nodeName, err := api.RPC.System.Name()
	if err != nil {
		panic(err)
	}
	nodeVersion, err := api.RPC.System.Version()
	if err != nil {
		panic(err)
	}

	fmt.Printf("You are connected to chain %v using %v v%v\n", chain, nodeName, nodeVersion)
}

func TestDecodeEvent(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI("ws://183.66.65.205:9944")
	meta, err := api.RPC.State.GetMetadataLatest()
	assert.NoError(t, err)
	numSet := []uint64{
		943,
		944,
		993,
		2143,
		2144,
		2193,
		3343,
		3344,
		3393,
		4543,
		4544,
		4593,
		5743,
		5744,
		5793,
		6943,
		6944,
		6993,
		8143,
		8144,
		8193,
		9343,
		9344,
		9393,
		10543,
		10544,
		10593,
		11742,
		11743,
		11792,
		12942,
		12943,
		12992,
		14142,
		14143,
		14192,
		15342,
		15343,
		15392,
		16542,
		16543,
		16592,
		17742,
		17743,
		17792,
		18942,
		18943,
		18992,
		20142,
		20143,
		20192,
		21342,
		21343,
		21392,
		22542,
		22543,
		22592,
		23742,
		23743,
		23792,
		24942,
		24943,
		24992,
		26142,
		26143,
		26192,
		27342,
		27343,
		27392,
		28061,
		28062,
		28542,
		28543,
		28592,
		29741,
		29742,
		29791,
		30941,
		30942,
		30991,
		32141,
		32142,
		32191,
		33341,
		33342,
		33391,
		34541,
		34542,
		34591,
		35741,
		35742,
		35791,
		36941,
		36942,
		36991,
		38141,
		38142,
		38191,
		39341,
		39342,
		39391,
		40541,
		40591,
		41741,
		41742,
		41791,
		42941,
		42942,
		42991,
		44141,
		44142,
		44191,
		45341,
		45342,
		45391,
		46541,
		46542,
		46591,
		47741,
		47742,
		47791,
		48941,
		48942,
		48991,
		50141,
		50142,
		50191,
		51341,
		51342,
		51391,
		52541,
		52542,
		52591,
		53741,
		53742,
		53791,
		54941,
		54942,
		54991,
		56141,
		56142,
		56191,
		57339,
		57340,
		57389,
		58539,
		58540,
		58589,
		59739,
		59740,
		59789,
		60939,
		60940,
		60989,
		62139,
		62140,
		62189,
		63339,
		63389,
		64539,
		64540,
		64589,
		65739,
		65740,
		65789,
		66939,
		66940,
		66989,
		68139,
		68140,
		68189,
		69339,
		69340,
		69389,
		70539,
		70540,
		70589,
		71739,
		71740,
		71789,
		72939,
		72940,
		72989,
		74139,
		74140,
		74189,
		75339,
		75340,
		75389,
		76539,
		76540,
		76589,
		77739,
		77740,
		77789,
		78939,
		78940,
		78989,
		80139,
		80140,
		80189,
		81339,
		81340,
		81389,
		82539,
		82540,
		82589,
		83739,
		83740,
		83789,
		84939,
		84940,
		84989,
		85331,
		85332,
		86139,
		86140,
		86189,
		87339,
		87340,
		87389,
		88539,
		88540,
		88589,
		89739,
		89740,
		89789,
		90939,
		90940,
		90989,
		92139,
		92140,
		92189,
		93339,
		93340,
		93389,
		94539,
		94540,
		94589,
		95739,
		95740,
		95789,
		96939,
		96940,
		96989,
		98139,
		98140,
		98189,
		99339,
		99340,
		99389,
		100539,
		100540,
		100589,
		101739,
		101740,
		101789,
		102939,
		102940,
		102989,
		104139,
		104140,
		104189,
		105339,
		105340,
		105389,
		106539,
		106540,
		106589,
		107739,
		107740,
		107789,
		108939,
		108940,
		108989,
		110139,
		110140,
		110189,
		111339,
		111340,
		111389,
		112538,
		112539,
		112588,
		113738,
		113739,
		113788,
		114938,
		114939,
		114988,
		116138,
		116139,
		116188,
		117338,
		117339,
		117388,
		118538,
		118539,
		118588,
		119738,
		119739,
		119788,
		120938,
		120939,
		120988,
		122138,
		122139,
		122188,
		123338,
		123339,
		123388,
		124538,
		124539,
		124588,
		125738,
		125739,
		125788,
		126938,
		126939,
		126988,
		128138,
		128139,
		128188,
		129338,
		129339,
		129388,
		130538,
		130539,
		130588,
		131738,
		131739,
		131788,
		132938,
		132939,
		132988,
		134138,
		134139,
		134188,
		135338,
		135339,
		135388,
		136538,
		136539,
		136588,
		137738,
		137739,
		137788,
		138938,
		138939,
		138988,
		140138,
		140139,
		140188,
		141338,
		141339,
		141388,
		142538,
		142539,
		142588,
		143106,
		143107,
		143738,
		143739,
		143788,
		//144015,
		144016,
		144938,
		144939,
		144988,
		146138,
		146139,
		146188,
		147338,
		147339,
		147388,
		148538,
		148539,
		148588,
		149738,
		149739,
		149788,
		150938,
		150939,
		150988,
		152137,
		152138,
		152187,
		153337,
		153338,
		153387,
		154537,
		154538,
		154587,
		155737,
		155738,
		155787,
		156937,
		156938,
		156987,
		158137,
		158138,
		158187,
		159337,
		159338,
		159387,
		160537,
		160538,
		160587,
		161737,
		161738,
		161787,
		162937,
		162938,
		162987,
		164137,
		164138,
		164187,
		165337,
		165338,
		165387,
		166537,
		166538,
		166587,
		167737,
		167738,
		167787,
		168937,
		168938,
		168987,
		170137,
		170138,
		170187,
		171337,
		171338,
		171387,
		172536,
		172537,
		172586,
		173736,
		173737,
		173786,
		174934,
		174935,
		174984,
		176134,
		176135,
		176184,
		177334,
		177335,
		177384,
		178534,
		178535,
		178584,
		179734,
		179735,
		179784,
		180934,
		180935,
		180984,
		182134,
		182135,
		182184,
		183334,
		183335,
		183384,
		184534,
		184535,
		184584,
		185734,
		185735,
		185784,
		186934,
		186935,
		186984,
		188134,
		188135,
		188184,
		189334,
		189335,
		189384,
		190534,
		190535,
		190584,
		191734,
	}
	numSet = []uint64{1762312}
	for _, num := range numSet {
		fmt.Println("num:", num)
		bh, err := api.RPC.Chain.GetBlockHash(num)
		assert.NoError(t, err)
		key, err := types.CreateStorageKey(meta, "System", "Events", nil)
		assert.NoError(t, err)
		raw, err := api.RPC.State.GetStorageRaw(key, bh)
		assert.NoError(t, err)
		// Decode the event records
		events := MyEventRecords{}

		err = DecodeEventRecordsWithIgnoreError(types.EventRecordsRaw(*raw), meta, &events)

		fmt.Println(events)
		if err != nil {
			log.Fatalln(err)
		}
	}

}
