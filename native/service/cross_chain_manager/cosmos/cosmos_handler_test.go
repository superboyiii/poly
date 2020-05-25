package cosmos

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/native"
	ccmcom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	synccom "github.com/ontio/multi-chain/native/service/header_sync/cosmos"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/ontio/multi-chain/native/storage"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var (
	acct *account.Account = account.NewAccount("")
)

const (
	SUCCESS = iota
	HEADER_NOT_EXIST
	PROOF_FORMAT_ERROR
	VERIFY_PROOT_ERROR
	TX_HAS_COMMIT
	UNKNOWN
)

func typeOfError(e error) int {
	if e == nil {
		return SUCCESS
	}
	errDesc := e.Error()
	if strings.Contains(errDesc, "GetHeaderByHeight, height is too big") {
		return HEADER_NOT_EXIST
	} else if strings.Contains(errDesc, "unmarshal proof error:") {
		return PROOF_FORMAT_ERROR
	} else if strings.Contains(errDesc, "verify proof value hash failed") {
		return VERIFY_PROOT_ERROR
	} else if strings.Contains(errDesc, "check done transaction error:checkDoneTx, tx already done") {
		return TX_HAS_COMMIT
	}
	return UNKNOWN
}

func NewNative(args []byte, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns := native.NewNativeService(db, nil, 0, 0, common.Uint256{0}, 0, args, false)

	contaractAddr, _ := hex.DecodeString("48A77F43C0D7A6D6f588c4758dbA22bf6C5D95a0")
	side := &side_chain_manager.SideChain{
		Name:         "cosmos",
		ChainId:     5,
		BlocksToWait: 1,
		Router:       1,
		CCMCAddress:  contaractAddr,
	}
	sink := common.NewZeroCopySink(nil)
	_ = side.Serialization(sink)
	ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(2)), cstates.GenRawStorageItem(sink.Bytes()))
	return ns
}

func TestProofHandle(t *testing.T) {
	cosmosHeaderSync := synccom.NewCosmosHandler()
	cosmosProofHandler := NewCosmosHandler()
	var native *native.NativeService
	{
		header10000, _ := hex.DecodeString("0aad020a02080a120a676169612d313330303718904e220b08cf86f8ef0510869ea13a305d3a480a201864d4fff0c86413801f6d00b8a42080311a4751e792daf1e301d5545986994e122408011220212ed39f6859a2ce4505fd78f12d18d07f5f52e868fca07eaaa380ed4d44404142204aae7c9169492bb1bd43b05b558db4b58e33b09938dae9b50cc8457f2341c10252209493d756ec5538cf367f1eb60ea607ce6b9b90bfaf5303c0c90f8d28c9b764ac5a209493d756ec5538cf367f1eb60ea607ce6b9b90bfaf5303c0c90f8d28c9b764ac62200f2908883a105c793b74495eb7d6df2eea479ed7fc9349206a65cb0f9987a0b86a208ec94a3cd9b68ace10067b95e43842b6ce599a54af402c7f3a6bbfa3a8965e2c820114099b2ec2e2adcdd37281ad383a2d51e437cfc92412bd130a480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd712b501080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510e3fba5e7023214099b2ec2e2adcdd37281ad383a2d51e437cfc92442401247ff95bf55c3083fcda1d936ba65c7f6aa7f5998902d059f834de016bac1a43e3756af63edea565bd2c31b9ca8a33da2dfdbb56f44ee7478d395b8a8acf10e12b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510a9f687820332141717093479fdf705e9defc4242321ff97acdc19d38014240e2f9bdce54f2ac25873ec3c420f4f8ebff8556b3eca884278ca67ede3ae1715a85d57a949d3103e7ed225466cfc81ea1591da72425fe73f9329af049437ef50d12b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d386f8ef0510c6b585e2023214193b773f29e934cddabf2b4b68ebfdd588a24678380242407d54374d4b3eb5866fb2c06f44d7a48d9f6ab47de78705aaf64a70b4cf92ead7aa1aacdaa3a4c522e9ae15b742015fa5d73029978ddb5cf19cb6065ecd267d0612b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef051091e8e08a0332141ef22e447b2ab74a268a9a4a80f0f512212bea6f3803424026d10598d99957a477ebfc1f22179415caf4e59caaf6f076d7d15445e73caeec3373cec7ade4e364d9c1540345cb190002c691ac37f8f47500836390d87b990112b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510b08897830332142b89383448acc1dd6b870385c16342f0f3950c31380442401dcec8996acbd1c546b39f5865ca85c5e65e9b2be949a61270e7953d823380b4dbeb1e914910e0bce3a6207d6a6a51885b7680e36a42d0c784d26576b74c250712b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510f5aee1e502321430486633a0aa2c19c8236e3a2e3f03c069320d8138054240b8ae34cba30732594ca70ac1aacfa437a40e8db7695749a8209d57f3fe9ae1e52e9c77a8d8fa1ae3b27094de45e63b5627a6334d9b43bee0d3e4aade9c6e1a0512b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510d6d6ec8703321432d45e3d9eb4aa86858203657d05fe16c4d617d338064240eb7b1edfd17840d3d71cfdcb9953cae78916de47d930db35dd72210ff201e7056b1bd85fcd42f33759d8c5c0d9ea04381ef4d24d6c911a31a00a4d97a4d89a0912b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510cfafc6ed02321442b5514cee989836e2a39af0f5184ced480923b538074240043db7a0ea4f5aef9291f3bff894443f4930b530a9e842a1e0b2fe5cdc1b3b3944e5b9fa22e39b1607cd4d27eb1643f062ea2cc198ae51da4d6072f668d5630a120012b601080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0b08d386f8ef051096ef914c3214a494f0968398acc7904a90d6970bd44ef4ce347f3809424059969595e8e8974731494188c2655c0cc98796f4a303db5cf1a8c6b003d8b76b70bbcffb8f4b46ffb6b99f6ef67a56561459a0b3dfe9c62a6b397d617cc0d70112b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510d285edec023214a68b3d8f585b920f0cfb084b1cb66b8a926f8907380a4240a3af8623098038d7ca5a656dfeeeeca19b32f6bca4825f4cc6b0f78ed89695de707dc755370b7f7aa5bf90e0d42bc9ca9326addd6cd04a6f4d231d9b8b31ff0312b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef05109695dff9023214b0febe5cc472b7fd8b97c55a226165523b8c80f5380b424024a3f51d6f2d45e9ef3dacee83b2d4d19c0e81bd7647ff23582ec694e5b5000d505c2ccc16cfd81c0733a3f486f0d8bca5fff8fc9b14623e7cf7f20bcaae5e0a120012b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef051087b9e782033214bc58b93fd5ec48da8e29afd353a87c0b4c94e9df380d424052554e7dae311733bd5c315ed289ce7003cf17951f5c6914b7f2ee2629dae77a362e6be9350c8b5d14e9977845e47417b0c5f2fc21d5936b250fd24f2c12dc0d12b701080210904e22480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd72a0c08d486f8ef0510c6e9e6e9023214ce5124e032e9a98d2a44e6d6c13fcba1ee2d443f380e4240b5fede2ec519deaa7ca238c0e8436fd8a11b8e19ef591d9ad6d633ca308880ca6db2c4343a29a579c79eb37d2bbe8ac2f1063cae5e33696d9fc08ddcdb739d0d1a4c0a14099b2ec2e2adcdd37281ad383a2d51e437cfc92412251624de64201b7f9f3bbfcce69aecedc371471d92158ed95924d68515f7db56b3bd0dbd1fca18d0860320919aeaffffffffffff011a4a0a141717093479fdf705e9defc4242321ff97acdc19d12251624de64206bee2aeba37718de65e92e73f06c56995eb982bf38da7c6f24d7579ad3da87131803209bf7d8ffffffffffff011a450a14193b773f29e934cddabf2b4b68ebfdd588a2467812251624de64206ec1bd2dc04db6259a0a7635ffa14eccb208faa902688d894683fd544c67856718d0860320bab80f1a450a141ef22e447b2ab74a268a9a4a80f0f512212bea6f12251624de64209fb70c212d71429f95f1e7ca3956e56793298b1dc10f98595459fec331a04ce818d0860320b9b80f1a450a142b89383448acc1dd6b870385c16342f0f3950c3112251624de64206d4a8dbaa9035f357c6efdf6b380086eff27a5f9cf4d919734445cbb9c98067618a58703208fba081a450a1430486633a0aa2c19c8236e3a2e3f03c069320d8112251624de64201d87fd2c0494d269e9c75be5918f320c4756c487a16d75555faea57dbe8493cb18d0860320b9b80f1a4c0a1432d45e3d9eb4aa86858203657d05fe16c4d617d312251624de6420d8ce09129ed710c9db0f823f660c20472e5949588f69add434e3dd25cede3e8018d0860320a59cf5ffffffffffff011a450a1442b5514cee989836e2a39af0f5184ced480923b512251624de6420acb885d1c6c018c37ae27ef3d7ab65301a2569b3009cafd4aab24537413c6d1d18d0860320bebe0d1a4c0a1498e21228648a20e7f3227cc93e89cf063223045912251624de6420ccc67b30eb8bfdc2ad194c1fee0fe4354218dac6db2ce9b58ad583d4a3106e8118d0860320d5bceaffffffffffff011a450a14a494f0968398acc7904a90d6970bd44ef4ce347f12251624de6420326eacb58b635ae7af46b156308fd8aa1beddd5b90752129be4d082d033cce6318d0860320b8b80f1a450a14a68b3d8f585b920f0cfb084b1cb66b8a926f890712251624de64205ab81acc09b40b4d09cabf28a8f270e750909858eae2cfe1c21f48074553966318d0860320d0b90c1a440a14b0febe5cc472b7fd8b97c55a226165523b8c80f512251624de6420329630f26416a5aa01cb2d985a63b464db44e60b41efab6a7d19420fa99651d318904e20cd850b1a4a0a14b7614527ab9650de12403896e1bad342bde6990c12251624de64208997c1ee9e0dafb8995c129060aef359400026f8946868d16c5fcf609299d794180120f38fdcffffffffffff011a450a14bc58b93fd5ec48da8e29afd353a87c0b4c94e9df12251624de64209a1cd7e2aabee8a564979e891d5c9c40ce7abb66e999f72a0388626bf3ea54b318d0860320b8b90c1a450a14ce5124e032e9a98d2a44e6d6c13fcba1ee2d443f12251624de64209b56c227c8477bb22802ea65b2a6738f2154ffec395cbfd5f35d0fa0510578e718d0860320e7b209")
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 5
		param.GenesisHeader = header10000
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), nil)
		err := cosmosHeaderSync.SyncGenesisHeader(native)
		if err != nil {
			fmt.Printf("err: %s", err.Error())
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
	{
		header10001, _ := hex.DecodeString("0aae020a02080a120a676169612d313330303718914e220c08d486f8ef0510d285edec02305d3a480a2002cb52df134dc60da7a5a17a46181f6e016ad274051afa0e5876a7d06e0666f7122408011220124f14738b6366089e27f7efd48533680e3a625fa4823bdddeb1e29844badbd74220315d2437192bae2bc606b040c6377908f294f51a1c826a000f6233f2cd2c583152209493d756ec5538cf367f1eb60ea607ce6b9b90bfaf5303c0c90f8d28c9b764ac5a209493d756ec5538cf367f1eb60ea607ce6b9b90bfaf5303c0c90f8d28c9b764ac62200f2908883a105c793b74495eb7d6df2eea479ed7fc9349206a65cb0f9987a0b86a20a271a678f37d0fae455698e3f2e59f2243c04d2dc922e0a6ed19c6c25337b88d820114193b773f29e934cddabf2b4b68ebfdd588a2467812be130a480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f4312b501080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef0510a8dce09d023214099b2ec2e2adcdd37281ad383a2d51e437cfc9244240b6e5364f9b14e010803180658f2ee2add16c9e63a554f48823f88ed3d5f9018036bb44bc510ed8b8f9bd4fac7a2e49cd5e422d7b3f130891d78abe1127a2ef0412b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef05109fe6b1ce0232141717093479fdf705e9defc4242321ff97acdc19d38014240ae01223bd2f5730321eccc2d883eb4295c30c84cec072571c21cf91853f8c0a9e77142cd9d143a9170fbe7e2b616797ce6760c1035da1f3338a6822d7d77d60e12b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08d986f8ef0510a68fe585033214193b773f29e934cddabf2b4b68ebfdd588a24678380242407853669f7804291e8b04e0e2c3cc4e4297c458d8dbb1c261d4e331bfa7252868f0aa62b705de5c1727a96dd2ac1bbb2b23c1cd727048b5b3d92443bc3adf7a0b12b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef051096ffa5bb0232141ef22e447b2ab74a268a9a4a80f0f512212bea6f380342404c771f4ffd2789ebb157597ed981fb3e2338b432596eec04f9c88614d2199629f1584cc3b2407abbbf2935f69148391b5a22d5bb40d55f9fa717c08a980e640d12b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef0510e5b78ca00232142b89383448acc1dd6b870385c16342f0f3950c3138044240c28c777bc51fd58f1d82b2a36ffc165c339dd0319568eb09f857a94e43776cba5067e5895be8f26fb95a71b660687d5ca30ef0feb2fe5144de33fb17c9190f0312b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef05108883a9d302321430486633a0aa2c19c8236e3a2e3f03c069320d8138054240fd9e5a1e28a49ffdf9883fdcc2e47e366d2986f2a6b50e8ec545b7be9e609d7f681765fde01236e14ff2017abf2ab9e33118aa668d3ac04edadeb6e1a6fffe0e12b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef0510ce9daea002321432d45e3d9eb4aa86858203657d05fe16c4d617d33806424087c2759a453157a56eea0f13c5ebc8e1af97d5df44640809e203525fdb37ea73f35b40f3fdb10abdbf5ea99a8f0d4dc70aff985195a6b56f14333ee5cb0c450e12b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef0510febda1cf02321442b5514cee989836e2a39af0f5184ced480923b538074240b3b1e4d05905118fd4ff176e535da718ddcbf27c59ec0748209b18f09b2da8fff614dd0ddbcfced998f36ac98b4d58bdd2b5e5faa3d14f1a51824460aebb580e120012b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08d886f8ef051085a7e1c6033214a494f0968398acc7904a90d6970bd44ef4ce347f38094240855cd0a041a8ac69fb4f6925880655398a35130a743fedfec99782275ea3ac6afac95a21704767f33a911daead867cecb28f9c00dd84de60682e6718aeeb870912b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef051086d8b5bc023214a68b3d8f585b920f0cfb084b1cb66b8a926f8907380a424041985840f4a4ffde1fe9ca83f36bf892a46de1a0e038343ed46ccd3040eaef789f1f59e1a02b7ab1b8e435da2c64d5ab40a58f809d642801e1936f8cfdc84d0a12b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef0510c5ffdfeb023214b0febe5cc472b7fd8b97c55a226165523b8c80f5380b4240bbdb5b18bf8f91fefac4a66be3a3b86ae4f8eef317a155cef4f6caa9c15869919ff191f76b1b38eb16b1fb05cd6c2745241d4a8f9abc26f64ec4f14461236804120012b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef0510d1c492ce023214bc58b93fd5ec48da8e29afd353a87c0b4c94e9df380d4240af3a58737b2513b7d670b964103778030b7ffde012812faf2c4340da138a4d40af8a7eb83b633cd215496939dc9c580514df5c80e1a15a4cef17aa8f6596460212b701080210914e22480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f432a0c08da86f8ef05108fbdecd4023214ce5124e032e9a98d2a44e6d6c13fcba1ee2d443f380e4240a8d56a8da870cf949e41d5ee030b1c6ee6a1d726866832430015b3ef3fc27e8bd710719dd0d61f44a110da9bb0dae0b8937c29f7e5773974066ec477f473110e1a4c0a14099b2ec2e2adcdd37281ad383a2d51e437cfc92412251624de64201b7f9f3bbfcce69aecedc371471d92158ed95924d68515f7db56b3bd0dbd1fca18d0860320e1a0edffffffffffff011a4a0a141717093479fdf705e9defc4242321ff97acdc19d12251624de64206bee2aeba37718de65e92e73f06c56995eb982bf38da7c6f24d7579ad3da87131803209ef7d8ffffffffffff011a4c0a14193b773f29e934cddabf2b4b68ebfdd588a2467812251624de64206ec1bd2dc04db6259a0a7635ffa14eccb208faa902688d894683fd544c67856718d0860320e1a0edffffffffffff011a450a141ef22e447b2ab74a268a9a4a80f0f512212bea6f12251624de64209fb70c212d71429f95f1e7ca3956e56793298b1dc10f98595459fec331a04ce818d086032089bf121a450a142b89383448acc1dd6b870385c16342f0f3950c3112251624de64206d4a8dbaa9035f357c6efdf6b380086eff27a5f9cf4d919734445cbb9c98067618a5870320b4c10b1a450a1430486633a0aa2c19c8236e3a2e3f03c069320d8112251624de64201d87fd2c0494d269e9c75be5918f320c4756c487a16d75555faea57dbe8493cb18d086032089bf121a4c0a1432d45e3d9eb4aa86858203657d05fe16c4d617d312251624de6420d8ce09129ed710c9db0f823f660c20472e5949588f69add434e3dd25cede3e8018d0860320f5a2f8ffffffffffff011a450a1442b5514cee989836e2a39af0f5184ced480923b512251624de6420acb885d1c6c018c37ae27ef3d7ab65301a2569b3009cafd4aab24537413c6d1d18d08603208ec5101a4c0a1498e21228648a20e7f3227cc93e89cf063223045912251624de6420ccc67b30eb8bfdc2ad194c1fee0fe4354218dac6db2ce9b58ad583d4a3106e8118d0860320a5c3edffffffffffff011a450a14a494f0968398acc7904a90d6970bd44ef4ce347f12251624de6420326eacb58b635ae7af46b156308fd8aa1beddd5b90752129be4d082d033cce6318d086032088bf121a450a14a68b3d8f585b920f0cfb084b1cb66b8a926f890712251624de64205ab81acc09b40b4d09cabf28a8f270e750909858eae2cfe1c21f48074553966318d0860320a0c00f1a440a14b0febe5cc472b7fd8b97c55a226165523b8c80f512251624de6420329630f26416a5aa01cb2d985a63b464db44e60b41efab6a7d19420fa99651d318904e20ddd30b1a4a0a14b7614527ab9650de12403896e1bad342bde6990c12251624de64208997c1ee9e0dafb8995c129060aef359400026f8946868d16c5fcf609299d794180120f48fdcffffffffffff011a450a14bc58b93fd5ec48da8e29afd353a87c0b4c94e9df12251624de64209a1cd7e2aabee8a564979e891d5c9c40ce7abb66e999f72a0388626bf3ea54b318d086032088c00f1a450a14ce5124e032e9a98d2a44e6d6c13fcba1ee2d443f12251624de64209b56c227c8477bb22802ea65b2a6738f2154ffec395cbfd5f35d0fa0510578e718d0860320b7b90c")
		//header10002, _ := hex.DecodeString("0aae020a02080a120a676169612d313330303718924e220c08da86f8ef051096ffa5bb02305d3a480a20dd73c370015d9aca8dbd7edea4d9e88da840b6818e23b4bd48fb32b74557e6ea122408011220f928830926270ed750a1ba920d1007126e02fbaca04c771fea03185fc42f8f434220e6524653dd757ef7f04e4c85fc1fe9dd625d8616566ddd5f7dddd71b0831394c52209493d756ec5538cf367f1eb60ea607ce6b9b90bfaf5303c0c90f8d28c9b764ac5a209493d756ec5538cf367f1eb60ea607ce6b9b90bfaf5303c0c90f8d28c9b764ac62200f2908883a105c793b74495eb7d6df2eea479ed7fc9349206a65cb0f9987a0b86a20cef6e566259c56925eceaad61fba4fcc882e71455f640d77802a410b058bd7588201141ef22e447b2ab74a268a9a4a80f0f512212bea6f12be130a480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda12b501080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef05108a949fbb023214099b2ec2e2adcdd37281ad383a2d51e437cfc9244240a2dfb64e39a356feaef523c2aa6e43a1b25e1b1ee12272ba52bfc7a7cfe073a901d698b23bc841fa230d41bf0ea652c31fa32282b0d471576a34b2bb037c830f12b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510eea1a0ba0232141717093479fdf705e9defc4242321ff97acdc19d380142403eaa1516237ee971f0e7b3fbe00a4b1a1e13ac60e93e10a1d4cdde8008a11e5b0479493b63f4738b19b566f0d34d3fb81da0be1063b4f64a69fa1b2d1ab5e70912b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08df86f8ef0510f6dfdf98023214193b773f29e934cddabf2b4b68ebfdd588a2467838024240fede9d1991bfb1bfd9c120c186e48c2b56a7925cc247415dbf0700e8cf3c256f5fcc4cd7716f6d3d06e8018d50d1fb5ce2b1f8cd46f0e36f19cd50871b359a0412b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510f6bcb5b60232141ef22e447b2ab74a268a9a4a80f0f512212bea6f3803424017c527fbabb0cc023792bf27fa438eef70aea1062e3b2621f9c0ba2dfda5ccedc7a4d0ba510469e5dd572ce422a460b764aa737e577eeada50b631f449c5c10c12b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510c8f0f2bc0232142b89383448acc1dd6b870385c16342f0f3950c3138044240b00b6ea1b52d5d5e0f3598adaeddd93f98957701d9f35b786d6bab86ce4d8070b2a03aa1116ff7f0e3afcb9a9393ed5d829d830cf4bba956b29fb7c22bf4570212b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef05109abcb8b302321430486633a0aa2c19c8236e3a2e3f03c069320d813805424092b03d77bc24f43086bf0e2c8cd6284bfa9e7e8b018053a3e7c14b2215967b56c6e983ef48e6639daf741d4335ae0541b371f6aa079834afde37082a5ad73f0212b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510ac899ac302321432d45e3d9eb4aa86858203657d05fe16c4d617d338064240a8f255a325b8a0e4396487fcbc6651ccb97b5f40621847ee0b8307bfde1ac335a11b3def8938729cd6a686d33c60bafba08f3e0e4f16d4f0b86ae036b302bd0f12b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510bbd0d8a402321442b5514cee989836e2a39af0f5184ced480923b538074240a99213790a0047db0e2effb6e850a0fe05bb0d4914467c4a3e5fe2588c5b9c8e6c988bcd17f886172ccb06c77670c2302bbdfb7a94b25a555c4edd1989a46801120012b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08de86f8ef0510d5d8e5ce033214a494f0968398acc7904a90d6970bd44ef4ce347f3809424070da06ad1f60b6be9abfbb85bd41a6657b852f6b499a0685e5e4bddfa59b675a71e982d3ccd04cdc13412eb886990fdf3c1736ebfa1944ffa5b705830d958c0b12b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510aacfd2a4023214a68b3d8f585b920f0cfb084b1cb66b8a926f8907380a42408d861dd1079695d1f38e4f26be6899fbb8ba04395747295b142418d003b15d1547d101b2ac1b130bf80b37dbe1f642066fb9f8020baf77a410f7e4735910a90912b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510e0efe2b6023214b0febe5cc472b7fd8b97c55a226165523b8c80f5380b424035fc43c593e7fdda458fe55632acb5ae64e03835341baa27f7c56b006cd472c687b7ca8ff347db45049f551665e8d68d0389a469b74293fa16cc2d9460fdf40a120012b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510c1f9b7b9023214bc58b93fd5ec48da8e29afd353a87c0b4c94e9df380d42402631554af8a08bd457024aa02dbed81a5d75bd707bb221c580257635342ce3c94f1c997bd625900fa2b1c50f268c13b5d24c90d8f79357f5216d4e96b797fb0912b701080210924e22480a2080d7695696bc5b2077265973dcf883c6c1bb67f7b4e6599da16a1cd218440c5a12240801122016caaca3754aac85490c039657af18e78489b25826372c23fdb4433cab435dda2a0c08e086f8ef0510cb9ed4a1023214ce5124e032e9a98d2a44e6d6c13fcba1ee2d443f380e424024723277c844f7b75395917f8ea888f82c6367636e455ed2fc06db08daa1a843cad55ee4f4a9951bbfaf22800a519f50494b0a8e4b797184a211693ccb8fa30b1a4c0a14099b2ec2e2adcdd37281ad383a2d51e437cfc92412251624de64201b7f9f3bbfcce69aecedc371471d92158ed95924d68515f7db56b3bd0dbd1fca18d0860320b1a7f0ffffffffffff011a4a0a141717093479fdf705e9defc4242321ff97acdc19d12251624de64206bee2aeba37718de65e92e73f06c56995eb982bf38da7c6f24d7579ad3da8713180320a1f7d8ffffffffffff011a4c0a14193b773f29e934cddabf2b4b68ebfdd588a2467812251624de64206ec1bd2dc04db6259a0a7635ffa14eccb208faa902688d894683fd544c67856718d0860320b1a7f0ffffffffffff011a4c0a141ef22e447b2ab74a268a9a4a80f0f512212bea6f12251624de64209fb70c212d71429f95f1e7ca3956e56793298b1dc10f98595459fec331a04ce818d0860320b0a7f0ffffffffffff011a450a142b89383448acc1dd6b870385c16342f0f3950c3112251624de64206d4a8dbaa9035f357c6efdf6b380086eff27a5f9cf4d919734445cbb9c98067618a5870320d9c80e1a450a1430486633a0aa2c19c8236e3a2e3f03c069320d8112251624de64201d87fd2c0494d269e9c75be5918f320c4756c487a16d75555faea57dbe8493cb18d0860320d9c5151a4c0a1432d45e3d9eb4aa86858203657d05fe16c4d617d312251624de6420d8ce09129ed710c9db0f823f660c20472e5949588f69add434e3dd25cede3e8018d0860320c5a9fbffffffffffff011a450a1442b5514cee989836e2a39af0f5184ced480923b512251624de6420acb885d1c6c018c37ae27ef3d7ab65301a2569b3009cafd4aab24537413c6d1d18d0860320decb131a4c0a1498e21228648a20e7f3227cc93e89cf063223045912251624de6420ccc67b30eb8bfdc2ad194c1fee0fe4354218dac6db2ce9b58ad583d4a3106e8118d0860320f5c9f0ffffffffffff011a450a14a494f0968398acc7904a90d6970bd44ef4ce347f12251624de6420326eacb58b635ae7af46b156308fd8aa1beddd5b90752129be4d082d033cce6318d0860320d8c5151a450a14a68b3d8f585b920f0cfb084b1cb66b8a926f890712251624de64205ab81acc09b40b4d09cabf28a8f270e750909858eae2cfe1c21f48074553966318d0860320f0c6121a440a14b0febe5cc472b7fd8b97c55a226165523b8c80f512251624de6420329630f26416a5aa01cb2d985a63b464db44e60b41efab6a7d19420fa99651d318904e20eda10c1a4a0a14b7614527ab9650de12403896e1bad342bde6990c12251624de64208997c1ee9e0dafb8995c129060aef359400026f8946868d16c5fcf609299d794180120f58fdcffffffffffff011a450a14bc58b93fd5ec48da8e29afd353a87c0b4c94e9df12251624de64209a1cd7e2aabee8a564979e891d5c9c40ce7abb66e999f72a0388626bf3ea54b318d0860320d8c6121a450a14ce5124e032e9a98d2a44e6d6c13fcba1ee2d443f12251624de64209b56c227c8477bb22802ea65b2a6738f2154ffec395cbfd5f35d0fa0510578e718d086032087c00f")
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 5
		param.Address = acct.Address
		param.Headers = append(param.Headers, header10001)
		//param.Headers = append(param.Headers, header10002)
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), native.GetCacheDB())
		err := cosmosHeaderSync.SyncBlockHeader(native)
		if err != nil {
			fmt.Printf("err: %s", err.Error())
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
	{
		param := new(ccmcom.EntranceParam)
		proof, _ := hex.DecodeString("0abc020a066961766c3a761215014de5e0db8c727e3f0bd34054c2ae5e450fd029721a9a0298020a95020a29080c102518904e2a20187abd33eaf95d44bcc547ae4d14685b497b3ab008f247f7922abb9e2354f6120a290808100d18864d222023d506a7e11b356f82afc811ea97fad934c7dafbf65f237620a56a91044eb8220a290806100718864d2220eb2bd0aa6cb2a42efa7a748483addeee4451f2ba6f392f7e53c6d377911832790a290804100418864d2a20bbc93f20c198a88dc498cf302dad2af3751d2d7d3a79605d4bc48f4f08e474400a290802100218864d2a20e7846c65a79c3de205ec66b3dc5455e02dfd4b2dcb2ae8e7b244d59d24c805fd1a3c0a15014de5e0db8c727e3f0bd34054c2ae5e450fd029721220e9062ffa7a56aefbdd9d8abe5627d7e8f90f04cfbc0cb0fe6ad8e31d1c20950618bd020ae3030a0a6d756c746973746f726512036163631acf03cd030aca030a2f0a046d696e7412270a2508904e122027f3839f1cbf6b64691decf2d0c53a25ed01c4701c8ad56b4e5c55d53f42ce360a2f0a046d61696e12270a2508904e122053b4300f5972c812ba19f6741d7ea102cd06852d648257521aefee565954e4790a320a077374616b696e6712270a2508904e1220a0ab3e8ca4cd98f6802576db6d1e09290dcd6fe93ef78f0d3c58a861daf281560a330a08736c617368696e6712270a2508904e1220f527b1da101472c4214f110a6508efb9cdb5c5c38138f22f654d28822c5d23ce0a310a06706172616d7312270a2508904e1220643156e9f056be28a099e51cb44c6e9a8339b4676b5079d571c7fd4bc9231be20a370a0c646973747269627574696f6e12270a2508904e1220b426a1501cf22d577ee3c859d962cf83fc010d60ee410e53b9670a53a628d3c90a2e0a03676f7612270a2508904e12204d5306f3ea08a49fbc6a8763323bcaf104c0c089f325fc57903e3349b5d5597d0a310a06737570706c7912270a2508904e12206ad1cc23efd842cb84fa8031912285cff38f35920cc513a876ef2360930b2d4e0a2e0a0361636312270a2508904e1220a2765987d658cb713211124a731fcd367b1a83ec380a600747f4c87fcd5e0ea0proof")
		value, _ := hex.DecodeString("0a362f6163632f2530314d254535254530254442253843727e253346253042254433405425433225414525354545253046254430253239721246f6e4f8380a144de5e0db8c727e3f0bd34054c2ae5e450fd029721a26eb5ae987210356ecd9c25d4565106c1c2e7d03e165c481bb7f55a291053ea7886964b7eec8ae20162801verify")
		param.SourceChainID = 5
		param.Height = 10001
		param.Proof = proof
		param.RelayerAddress = acct.Address[:]
		param.Extra = value
		param.HeaderOrCrossChainMsg = []byte{}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), native.GetCacheDB())
		_, err := cosmosProofHandler.MakeDepositProposal(native)
		if err != nil {
			fmt.Printf("%v", err)
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
}