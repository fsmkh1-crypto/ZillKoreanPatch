// SPDX-License-Identifier: GPL-3.0-or-later

package message

// KoreanConsumerStorageText returns the build-owned compact wording required by
// proven fixed-size retail consumers. These are semantic translations, not
// layout rewrites: each override keeps the source meaning while fitting a
// consumer whose storage ceiling cannot be repaired with line breaks.
//
// Keep this as the single projection seam used by both materialization and the
// repository-wide storage-contract audit. The legacy character-choice map is
// populated here as well because materialization already consumes that map.
func KoreanConsumerStorageText(id int, canonical string) string {
	if text, ok := characterChoiceBufferDiagnostic[id]; ok {
		return text
	}
	return canonical
}

func init() {
	// Character-creation choices: fixed 31-byte field including terminator.
	characterChoiceBufferDiagnostic[10016] = "약점을 찾아 여러 방법 시도<end>"
	characterChoiceBufferDiagnostic[10017] = "힘을 믿고 정면 싸운다<end>"
	characterChoiceBufferDiagnostic[10020] = "초원의 아름다움 그림에 담는다<end>"
	characterChoiceBufferDiagnostic[10025] = "강인한 의지의 늠름한 표정<end>"
	characterChoiceBufferDiagnostic[10026] = "모든 것을 감싸는 온화한 표정<end>"
	characterChoiceBufferDiagnostic[10034] = "폭발적 파괴력을 내는 체력<end>"
	characterChoiceBufferDiagnostic[10071] = "물살을 거슬러 오르는 물고기<end>"

	// Bounded labels: fixed 28-byte field including terminator.
	characterChoiceBufferDiagnostic[1670002] = "네메아를 좋아하는 아저씨<end>"
	characterChoiceBufferDiagnostic[1670033] = "모험자 싫어하는 왕궁 문지기<end>"
	characterChoiceBufferDiagnostic[1670083] = "사막 민족에 관심 있는 남자<end>"
	characterChoiceBufferDiagnostic[1670289] = "겁 없는 배 지킴이<end>"

	// Guild-region descriptions: fixed 152-byte field including terminator.
	characterChoiceBufferDiagnostic[160053] = "유사 이전부터 있던 탑. 최상층 제단의 용도는 불명이다. 바람 잘 통하는 곳에 풍정의 결정이 있고, 풍여조 둥지에는 풍여조의 깃털이 있다.<end>"
	characterChoiceBufferDiagnostic[160055] = "중앙에 다섯 기둥이 선 작은 섬이 있는 호수. 근처 폭포에서 타레몰게의 기수를 채취할 수 있다. 북쪽 기슭의 세즈네 독초에서 세즈네의 뿌리를 채취할 수 있다.<end>"
	characterChoiceBufferDiagnostic[160065] = "대륙 최대 광산도 지금은 몬스터 소굴이다. 드물게 연강석이 떨어진다. 하얗게 빛나는 향데광을 채취할 수 있고, 나락의 구멍 근처에는 토정의 결정이 있다.<end>"
	characterChoiceBufferDiagnostic[160073] = "용왕이 산다는 화산섬. 동쪽 바위 지대에 평평한 오포스의 바위가 있고 타레몰게의 기수도 솟는다. 어딘가에 노로가메 서식지가 있다고 한다.<end>"

	// Trap help: fixed 104-byte field including terminator; both substitutions
	// and color controls remain in their original order.
	characterChoiceBufferDiagnostic[1070079] = "난이도 <color:e><value:$1A><color:f> 함정. 해제에는 DEX <color:e><value:$1B><color:f> 이상이 필요합니다.<end>"
}
