package consensus

import (
	"encoding/hex"
	"testing"

	"github.com/threefoldtech/rivine/encoding"
)

func TestDecodeLegacyProcessedBlock(t *testing.T) {
	testCases := []string{
		// {
		//   "Block": {
		//     "ParentID": "640be38bb81485ec714fa0c5ab851abbd76c9d37e3c841d8d48feb9edce0470d",
		//     "Timestamp": 1523184488,
		//     "POBSOutput": {
		//       "BlockHeight": 3348,
		//       "TransactionIndex": 0,
		//       "OutputIndex": 0
		//     },
		//     "MinerPayouts": [
		//       {
		//         "value": "10000000000",
		//         "unlockhash": "0188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf43a100d570718"
		//       }
		//     ],
		//     "Transactions": [
		//       {
		//         "version": 0,
		//         "data": {
		//           "coininputs": [],
		//           "blockstakeinputs": [
		//             {
		//               "parentid": "441deb06592694fb0f7fb41dd9ec7239b01cf83ec0100f92c5b3312ebf5e9fec",
		//               "unlocker": {
		//                 "type": 1,
		//                 "condition": {
		//                   "publickey": "ed25519:bed5562f68ae34afc4c4b8e4bf7768627329cf9dfb87072963c662e18aa736ff"
		//                 },
		//                 "fulfillment": {
		//                   "signature": "81213c5f24700866b6ae255e2fd41f385d80f416f11c6166c1944ebf3ecc5d070e53d54a2882f49febf587ecc4f6a8bc1d220d4a61a91a9678f9049acd9b7c00"
		//                 }
		//               }
		//             }
		//           ],
		//           "blockstakeoutputs": [
		//             {
		//               "value": "1000",
		//               "unlockhash": "0188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf43a100d570718"
		//             }
		//           ],
		//           "minerfees": null
		//         }
		//       }
		//     ]
		//   },
		//   "Height": 3349,
		//   "Depth": [
		//     0, 0, 0, 4, 203, 162, 109, 77, 53, 59, 207, 2, 25,
		//     244, 240, 37, 127, 129, 126, 144, 216, 134, 79, 22, 164, 5,
		//     143, 214, 189, 91, 112, 27
		//   ],
		//   "ChildTarget": [
		//     0, 0, 45, 155, 207, 85, 172, 223, 126, 185, 29, 189, 99, 47,
		//     165, 244, 65, 226, 254, 104, 165, 146, 112, 74, 60, 9, 65, 155,
		//     130, 148, 185, 50
		//   ],
		//   "DiffsGenerated": true,
		//   "CoinOutputDiffs": [
		//     {
		//       "Direction": true,
		//       "ID": "b530693498b9968cf78c0a3e796d573c64c6989a05f698964c5d9d6b2e09da45",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01e34588bee49b2cbd53f2198cd5022fbbe78aecb8125a39efb8699720b946e84ead718daf0cd6"
		//       }
		//     }
		//   ],
		//   "BlockStakeOutputDiffs": [
		//     {
		//       "Direction": false,
		//       "ID": "441deb06592694fb0f7fb41dd9ec7239b01cf83ec0100f92c5b3312ebf5e9fec",
		//       "BlockStakeOutput": {
		//         "Value": "1000",
		//         "UnlockHash": "0188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf43a100d570718"
		//       }
		//     },
		//     {
		//       "Direction": true,
		//       "ID": "9caaf5e1ed6aa1bce252aba995d7e14c82005f3da15bf08a8425d9c86eedafcf",
		//       "BlockStakeOutput": {
		//         "Value": "1000",
		//         "UnlockHash": "0188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf43a100d570718"
		//       }
		//     }
		//   ],
		//   "DelayedCoinOutputDiffs": [
		//     {
		//       "Direction": true,
		//       "ID": "707a8d2f97182646caeb2404e5fdd59119a3d2de4d2f40efbde4e3dca8f7d132",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "0188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf43a100d570718"
		//       },
		//       "MaturityHeight": 4069
		//     },
		//     {
		//       "Direction": false,
		//       "ID": "b530693498b9968cf78c0a3e796d573c64c6989a05f698964c5d9d6b2e09da45",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01e34588bee49b2cbd53f2198cd5022fbbe78aecb8125a39efb8699720b946e84ead718daf0cd6"
		//       },
		//       "MaturityHeight": 3349
		//     }
		//   ],
		//   "TxIDDiffs": [
		//     {
		//       "Direction": true,
		//       "LongID": "ec6620d8c7813cc86ae5c6021d4bb27f0fd77b4f8a4d98f2844032e64d071f4f",
		//       "ShortID": 54870016
		//     }
		//   ],
		//   "ConsensusChecksum": "beeb48fb26bd7cd43e97d0a1a974241b5c071bea3c736fec07d6685da6e39213"
		// }
		`640be38bb81485ec714fa0c5ab851abbd76c9d37e3c841d8d48feb9edce0470d68f3c95a00000000140d000000000000000000000000000000000000000000000100000000000000050000000000000002540be4000188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4010000000000000000000000000000000000000000000000000100000000000000441deb06592694fb0f7fb41dd9ec7239b01cf83ec0100f92c5b3312ebf5e9fec013800000000000000656432353531390000000000000000002000000000000000bed5562f68ae34afc4c4b8e4bf7768627329cf9dfb87072963c662e18aa736ff400000000000000081213c5f24700866b6ae255e2fd41f385d80f416f11c6166c1944ebf3ecc5d070e53d54a2882f49febf587ecc4f6a8bc1d220d4a61a91a9678f9049acd9b7c000100000000000000020000000000000003e80188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf400000000000000000000000000000000150d00000000000000000004cba26d4d353bcf0219f4f0257f817e90d8864f16a4058fd6bd5b701b00002d9bcf55acdf7eb91dbd632fa5f441e2fe68a592704a3c09419b8294b93201010000000000000001b530693498b9968cf78c0a3e796d573c64c6989a05f698964c5d9d6b2e09da45050000000000000002540be40001e34588bee49b2cbd53f2198cd5022fbbe78aecb8125a39efb8699720b946e84e020000000000000000441deb06592694fb0f7fb41dd9ec7239b01cf83ec0100f92c5b3312ebf5e9fec020000000000000003e80188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4019caaf5e1ed6aa1bce252aba995d7e14c82005f3da15bf08a8425d9c86eedafcf020000000000000003e80188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4020000000000000001707a8d2f97182646caeb2404e5fdd59119a3d2de4d2f40efbde4e3dca8f7d132050000000000000002540be4000188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4e50f00000000000000b530693498b9968cf78c0a3e796d573c64c6989a05f698964c5d9d6b2e09da45050000000000000002540be40001e34588bee49b2cbd53f2198cd5022fbbe78aecb8125a39efb8699720b946e84e150d000000000000010000000000000001ec6620d8c7813cc86ae5c6021d4bb27f0fd77b4f8a4d98f2844032e64d071f4f0040450300000000beeb48fb26bd7cd43e97d0a1a974241b5c071bea3c736fec07d6685da6e39213`,
		// {
		//   "Block": {
		//     "ParentID": "e98c68c24b253d6f162e33504694ad929fac7862c2287adeffbd2a7b7062f126",
		//     "Timestamp": 1523890287,
		//     "POBSOutput": {
		//       "BlockHeight": 8809,
		//       "TransactionIndex": 0,
		//       "OutputIndex": 0
		//     },
		//     "MinerPayouts": [
		//       {
		//         "value": "10000000000",
		//         "unlockhash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     ],
		//     "Transactions": [
		//       {
		//         "version": 0,
		//         "data": {
		//           "coininputs": [],
		//           "blockstakeinputs": [
		//             {
		//               "parentid": "bcd60b89fe5acc48ff45256d857ac479804a35c13f0ab1ea5df5cfd58936bd40",
		//               "unlocker": {
		//                 "type": 1,
		//                 "condition": {
		//                   "publickey": "ed25519:fe385d16409b9aba8828d5f3431229d624be83f9a98b93ffab88c1d10d3dcae8"
		//                 },
		//                 "fulfillment": {
		//                   "signature": "60287949a71dd9f9702328f37ecab13cc84296471ba8a0c417a3d96919581b3b17472fb229369fda3bf6399af7598d8f90571cfe76feeab7102ccc38b89be403"
		//                 }
		//               }
		//             }
		//           ],
		//           "blockstakeoutputs": [
		//             {
		//               "value": "900",
		//               "unlockhash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//             }
		//           ],
		//           "minerfees": null
		//         }
		//       }
		//     ]
		//   },
		//   "Height": 8810,
		//   "Depth": [
		//     0, 0, 0, 2, 191, 214, 104, 159, 114, 126, 234, 255, 143,
		//     146, 32, 151, 99, 3, 212, 243, 76, 93, 73, 94, 76, 172,
		//     206, 21, 245, 50, 174, 206
		//   ],
		//   "ChildTarget": [
		//     0, 0, 157, 71, 45, 179, 179, 205, 68, 103, 123, 209, 169,
		//     45, 240, 119, 19, 58, 225, 133, 119, 141, 58, 179, 36, 242,
		//     108, 239, 45, 18, 196, 132
		//   ],
		//   "DiffsGenerated": true,
		//   "CoinOutputDiffs": [
		//     {
		//       "Direction": true,
		//       "ID": "f4ad9e4fdf7b57d074072c94fff585797678ac9272a466ca08be4178da671bff",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     }
		//   ],
		//   "BlockStakeOutputDiffs": [
		//     {
		//       "Direction": false,
		//       "ID": "bcd60b89fe5acc48ff45256d857ac479804a35c13f0ab1ea5df5cfd58936bd40",
		//       "BlockStakeOutput": {
		//         "Value": "900",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     },
		//     {
		//       "Direction": true,
		//       "ID": "f05b36029043b468031ac34b6b84d762a879b2d061aad8a7a1e84043b9e38d70",
		//       "BlockStakeOutput": {
		//         "Value": "900",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     }
		//   ],
		//   "DelayedCoinOutputDiffs": [
		//     {
		//       "Direction": true,
		//       "ID": "75bee4e9377be4abb66fba623a295bba8836dfc35bb85d4def388ad8ea21574b",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       },
		//       "MaturityHeight": 9530
		//     },
		//     {
		//       "Direction": false,
		//       "ID": "f4ad9e4fdf7b57d074072c94fff585797678ac9272a466ca08be4178da671bff",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       },
		//       "MaturityHeight": 8810
		//     }
		//   ],
		//   "TxIDDiffs": [
		//     {
		//       "Direction": true,
		//       "LongID": "6a0d32e7335c674aceb34505d8f36695956a512eeed83a472ada34ac214ce0d0",
		//       "ShortID": 144343040
		//     }
		//   ],
		//   "ConsensusChecksum": "40c0bd83fb9e44759eaf52b5d451820bd92e2b3cd076689a0891910c34325f4b"
		// }
		`e98c68c24b253d6f162e33504694ad929fac7862c2287adeffbd2a7b7062f1266fb8d45a000000006922000000000000000000000000000000000000000000000100000000000000050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9010000000000000000000000000000000000000000000000000100000000000000bcd60b89fe5acc48ff45256d857ac479804a35c13f0ab1ea5df5cfd58936bd40013800000000000000656432353531390000000000000000002000000000000000fe385d16409b9aba8828d5f3431229d624be83f9a98b93ffab88c1d10d3dcae8400000000000000060287949a71dd9f9702328f37ecab13cc84296471ba8a0c417a3d96919581b3b17472fb229369fda3bf6399af7598d8f90571cfe76feeab7102ccc38b89be40301000000000000000200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9000000000000000000000000000000006a2200000000000000000002bfd6689f727eeaff8f9220976303d4f34c5d495e4cacce15f532aece00009d472db3b3cd44677bd1a92df077133ae185778d3ab324f26cef2d12c48401010000000000000001f4ad9e4fdf7b57d074072c94fff585797678ac9272a466ca08be4178da671bff050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9020000000000000000bcd60b89fe5acc48ff45256d857ac479804a35c13f0ab1ea5df5cfd58936bd400200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d901f05b36029043b468031ac34b6b84d762a879b2d061aad8a7a1e84043b9e38d700200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d902000000000000000175bee4e9377be4abb66fba623a295bba8836dfc35bb85d4def388ad8ea21574b050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d93a2500000000000000f4ad9e4fdf7b57d074072c94fff585797678ac9272a466ca08be4178da671bff050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d96a220000000000000100000000000000016a0d32e7335c674aceb34505d8f36695956a512eeed83a472ada34ac214ce0d000809a080000000040c0bd83fb9e44759eaf52b5d451820bd92e2b3cd076689a0891910c34325f4b`,
		// {
		//   "Block": {
		//     "ParentID": "855497fce57518bc896634db32ee9dff5cd900dbe5ecd62c6ba5f1110ebde4e1",
		//     "Timestamp": 1523679203,
		//     "POBSOutput": {
		//       "BlockHeight": 7047,
		//       "TransactionIndex": 0,
		//       "OutputIndex": 0
		//     },
		//     "MinerPayouts": [
		//       {
		//         "value": "10000000000",
		//         "unlockhash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     ],
		//     "Transactions": [
		//       {
		//         "version": 0,
		//         "data": {
		//           "coininputs": [],
		//           "blockstakeinputs": [
		//             {
		//               "parentid": "c035da27a2d52530e7e121a9f9149aad29cce9bf2b13567ffdaa83fc94a43ccd",
		//               "unlocker": {
		//                 "type": 1,
		//                 "condition": {
		//                   "publickey": "ed25519:fe385d16409b9aba8828d5f3431229d624be83f9a98b93ffab88c1d10d3dcae8"
		//                 },
		//                 "fulfillment": {
		//                   "signature": "ffcc85d99787a7e6d8c5671cc856e3f71587e2b3e551b34477299cca5e1353478a133c9565d611e7d92b66f5b2e7827258d79aa39a6afe6bc877a5d133bad40d"
		//                 }
		//               }
		//             }
		//           ],
		//           "blockstakeoutputs": [
		//             {
		//               "value": "900",
		//               "unlockhash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//             }
		//           ],
		//           "minerfees": null
		//         }
		//       }
		//     ]
		//   },
		//   "Height": 7048,
		//   "Depth": [
		//     0, 0, 0, 3, 29, 108, 174, 160, 196, 131, 157, 233, 143,
		//     43, 23, 216, 179, 143, 18, 153, 21, 73, 208, 244, 55, 215,
		//     129, 70, 127, 177, 89, 17
		//   ],
		//   "ChildTarget": [
		//     0, 0, 163, 132, 107, 218, 212, 95, 234, 58, 230, 133, 42,
		//     159, 239, 45, 251, 154, 191, 117, 131, 99, 205, 41, 120, 187,
		//     174, 141, 182, 81, 4, 248
		//   ],
		//   "DiffsGenerated": true,
		//   "CoinOutputDiffs": [
		//     {
		//       "Direction": true,
		//       "ID": "567993c3b1385771288d8d626ab1203c84ff55daada37ce2e5eef8998f59aa88",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     }
		//   ],
		//   "BlockStakeOutputDiffs": [
		//     {
		//       "Direction": false,
		//       "ID": "c035da27a2d52530e7e121a9f9149aad29cce9bf2b13567ffdaa83fc94a43ccd",
		//       "BlockStakeOutput": {
		//         "Value": "900",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     },
		//     {
		//       "Direction": true,
		//       "ID": "a8091aa751ee717b353cae5976ff3d59f476807d9975c2bb213c92e1b7fea175",
		//       "BlockStakeOutput": {
		//         "Value": "900",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       }
		//     }
		//   ],
		//   "DelayedCoinOutputDiffs": [
		//     {
		//       "Direction": true,
		//       "ID": "21e7b7a3fa208da35c7d8eb2510ece2545a71f4c6190997e576bedcd590ef99c",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       },
		//       "MaturityHeight": 7768
		//     },
		//     {
		//       "Direction": false,
		//       "ID": "567993c3b1385771288d8d626ab1203c84ff55daada37ce2e5eef8998f59aa88",
		//       "CoinOutput": {
		//         "Value": "10000000000",
		//         "UnlockHash": "01ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9d47e61e0261f"
		//       },
		//       "MaturityHeight": 7048
		//     }
		//   ],
		//   "TxIDDiffs": [
		//     {
		//       "Direction": true,
		//       "LongID": "ccac3be2005865f4f15f6934e4b7993ff1d3d584737e6bdda6adca30af6acd8e",
		//       "ShortID": 115474432
		//     }
		//   ],
		//   "ConsensusChecksum": "14e65d61367ce2195150a8c9e9a670d42f52f82897c798a515afa384fdffa371"
		// }
		`855497fce57518bc896634db32ee9dff5cd900dbe5ecd62c6ba5f1110ebde4e1e37fd15a00000000871b000000000000000000000000000000000000000000000100000000000000050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9010000000000000000000000000000000000000000000000000100000000000000c035da27a2d52530e7e121a9f9149aad29cce9bf2b13567ffdaa83fc94a43ccd013800000000000000656432353531390000000000000000002000000000000000fe385d16409b9aba8828d5f3431229d624be83f9a98b93ffab88c1d10d3dcae84000000000000000ffcc85d99787a7e6d8c5671cc856e3f71587e2b3e551b34477299cca5e1353478a133c9565d611e7d92b66f5b2e7827258d79aa39a6afe6bc877a5d133bad40d01000000000000000200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d900000000000000000000000000000000881b000000000000000000031d6caea0c4839de98f2b17d8b38f12991549d0f437d781467fb159110000a3846bdad45fea3ae6852a9fef2dfb9abf758363cd2978bbae8db65104f801010000000000000001567993c3b1385771288d8d626ab1203c84ff55daada37ce2e5eef8998f59aa88050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9020000000000000000c035da27a2d52530e7e121a9f9149aad29cce9bf2b13567ffdaa83fc94a43ccd0200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d901a8091aa751ee717b353cae5976ff3d59f476807d9975c2bb213c92e1b7fea1750200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d902000000000000000121e7b7a3fa208da35c7d8eb2510ece2545a71f4c6190997e576bedcd590ef99c050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9581e00000000000000567993c3b1385771288d8d626ab1203c84ff55daada37ce2e5eef8998f59aa88050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9881b000000000000010000000000000001ccac3be2005865f4f15f6934e4b7993ff1d3d584737e6bdda6adca30af6acd8e0000e2060000000014e65d61367ce2195150a8c9e9a670d42f52f82897c798a515afa384fdffa371`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var lpb legacyProcessedBlock
		err = encoding.Unmarshal(b, &lpb)
		if err != nil {
			t.Error(idx, err)
		}
	}
}

func TestDecodeLegacyCoinOutput(t *testing.T) {
	testCases := []string{
		`050000000000000002540be4000188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4`,
		`050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var output legacyOutput
		err = encoding.Unmarshal(b, &output)
		if err != nil {
			t.Error(idx, err)
		}
	}
}

func TestDecodeLegacyBlockStakeOutput(t *testing.T) {
	testCases := []string{
		`0200000000000000038401ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9`,
		`020000000000000003e801e34588bee49b2cbd53f2198cd5022fbbe78aecb8125a39efb8699720b946e84e`,
		`020000000000000003e801bd7048f40168df7d837fd398bcffdf2d69d992ef53bd1677570d373ba378edea`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var output legacyOutput
		err = encoding.Unmarshal(b, &output)
		if err != nil {
			t.Error(idx, err)
		}
	}
}
