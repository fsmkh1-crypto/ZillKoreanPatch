#!/usr/bin/env python3
"""Emit high-confidence exact Korean translations from a work packet.

This intentionally handles only context-free literals whose Korean rendering is
stable. Everything else remains for LLM translation. Runtime control tokens are
kept verbatim by mapping the complete canonical string.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path

EXACT = {
    "未使用<end>": "미사용<end>",
    "<未使用><end>": "<미사용><end>",
    "＜無＞<end>": "＜없음＞<end>",
    "終了<end>": "종료<end>",
    "　予備<end>": "　예비<end>",
    "予備<end>": "예비<end>",
    "宿屋<end>": "여관<end>",
    " 魔   法<end>": " 마   법<end>",
    " 能力値<end>": " 능력치<end>",
    " 装   備<end>": " 장   비<end>",
    " スキル<end>": " 스킬<end>",
    " ソウル<end>": " 소울<end>",
    " 情   報<end>": " 정   보<end>",

    # Stable standalone character/race/system labels.
    "男<end>": "남성<end>",
    "女<end>": "여성<end>",
    "なし<end>": "없음<end>",
    "人間<end>": "인간<end>",
    "エルフ<end>": "엘프<end>",
    "コーンス<end>": "콘스<end>",
    "ダークエルフ<end>": "다크 엘프<end>",
    "ドワーフ<end>": "드워프<end>",
    "ハイエルフ<end>": "하이 엘프<end>",
    "魔人<end>": "마인<end>",
    "ゴブリン<end>": "고블린<end>",
    "最大<end>": "최대<end>",
    "毒<end>": "독<end>",
    "電気<end>": "전기<end>",
    "冷気<end>": "냉기<end>",
    "光<end>": "빛<end>",
    "爆発<end>": "폭발<end>",
    "退治<end>": "퇴치<end>",
    "探索<end>": "탐색<end>",
    "配達<end>": "배달<end>",
    "護衛<end>": "호위<end>",
    "救出<end>": "구출<end>",
    "美しさ<end>": "아름다움<end>",
    "復讐？<end>": "복수?<end>",
    "聖杯？<end>": "성배?<end>",
    "青竜将軍<end>": "청룡장군<end>",
    "白虎将軍<end>": "백호장군<end>",
    "朱雀将軍<end>": "주작장군<end>",
    "玄武将軍<end>": "현무장군<end>",
    "水の巫女<end>": "물의 무녀<end>",
    "風の巫女<end>": "바람의 무녀<end>",
    "地の巫女<end>": "땅의 무녀<end>",
    "救世主<end>": "구세주<end>",

    # Stable standalone menu/UI labels.
    "アイテム<end>": "아이템<end>",
    "システム<end>": "시스템<end>",
    "メニュー<end>": "메뉴<end>",
    "キャンセル<end>": "취소<end>",
    "イベント<end>": "이벤트<end>",
    "表示切替<end>": "표시 전환<end>",
    "戦闘開始<end>": "전투 시작<end>",
    "種類選択<end>": "종류 선택<end>",
    "対象選択<end>": "대상 선택<end>",
    "隊列変更<end>": "진형 변경<end>",
    "人物切替<end>": "인물 전환<end>",
    "全身表示<end>": "전신 표시<end>",
    "通信機<end>": "통신기<end>",
    "予選大会<end>": "예선대회<end>",
    "決勝大会<end>": "결승대회<end>",
    "小刀手斧<end>": "단검·손도끼<end>",
    "長老の家<end>": "장로의 집<end>",
    "領主邸前<end>": "영주 저택 앞<end>",
    "だみぃ2_7<end>": "더미2_7<end>",
    "だみぃ2_8<end>": "더미2_8<end>",
    "説明不要<end>": "설명 불필요<end>",
    "無視する<end>": "무시한다<end>",
    "根拠は？<end>": "근거는?<end>",

    # Stable standalone NPC/job labels and proper names already established in
    # the Korean corpus.
    "冒険者<end>": "모험가<end>",
    "傭兵<end>": "용병<end>",
    "魔道士<end>": "마도사<end>",
    "吟遊詩人<end>": "음유시인<end>",
    "巡礼者<end>": "순례자<end>",
    "旅行者<end>": "여행자<end>",
    "司祭<end>": "사제<end>",
    "士官<end>": "장교<end>",
    "裁判官<end>": "재판관<end>",
    "執事<end>": "집사<end>",
    "使用人<end>": "하인<end>",
    "商人<end>": "상인<end>",
    "料理人<end>": "요리사<end>",
    "学生<end>": "학생<end>",
    "船員<end>": "선원<end>",
    "修道女<end>": "수녀<end>",
    "衛兵<end>": "경비병<end>",
    "研究生<end>": "연구생<end>",
    "女学生<end>": "여학생<end>",
    "女の子<end>": "여자아이<end>",
    "男の子<end>": "남자아이<end>",
    "受付嬢<end>": "접수원<end>",
    "警備兵<end>": "경비병<end>",
    "貿易商<end>": "무역상<end>",
    "探検家<end>": "탐험가<end>",
    "宿泊客<end>": "숙박객<end>",
    "常連客<end>": "단골손님<end>",
    "見物客<end>": "구경꾼<end>",
    "団長！<end>": "단장!<end>",
    "親分！<end>": "두목!<end>",
    "師匠？<end>": "스승님?<end>",
    "セラ？<end>": "세라?<end>",
    "ママ！<end>": "엄마!<end>",
    "ヴァン！<end>": "반!<end>",

    # Plain/casual Japanese forms already carry register.  These mappings are
    # limited to rows whose core lexical meaning is also stable enough to make
    # the default Korean banmal deterministic.  Ambiguous rows such as いいよ
    # still stay out of EXACT and are translated with scene context.
    "誰だ？<end>": "누구야?<end>",
    "待て。<end>": "기다려.<end>",
    "待て！<end>": "기다려!<end>",
    "任せて<end>": "맡겨<end>",
    "忙しい<end>": "바빠<end>",
    "別に…<end>": "별로…<end>",
    "行け！<end>": "가!<end>",
    "いやだ<end>": "싫어<end>",
    "殺せ！<end>": "죽여!<end>",
    "なに？<end>": "뭐?<end>",
    "大丈夫<end>": "괜찮아<end>",
    "見たい<end>": "보고 싶어<end>",
    "俺は…<end>": "나는…<end>",
    "任せて！<end>": "맡겨!<end>",
    "死ね！<end>": "죽어!<end>",
    "聞いた<end>": "들었어<end>",
    "暇です<end>": "한가합니다<end>",
    "見える<end>": "보여<end>",
    "言える<end>": "말할 수 있어<end>",
    "誓う。<end>": "맹세해.<end>",
    "寒い…。<end>": "추워…<end>",
    "撃て！！<end>": "쏴!!<end>",
    "加わる<end>": "합류한다<end>",
    "逃げる<end>": "도망친다<end>",
    "助ける<end>": "돕는다<end>",
    "見守る<end>": "지켜본다<end>",
    "失礼。<end>": "실례.<end>",
    "喜んで<end>": "기꺼이.<end>",
    "みんな？<end>": "다들?<end>",
    "…罠か。<end>": "…함정인가.<end>",
    "戻るぞ！<end>": "돌아가자!<end>",
    "…来い。<end>": "…와.<end>",
    "…よし！<end>": "…좋아!<end>",
    "もういい<end>": "됐어<end>",
    "もちろん<end>": "물론이지<end>",
    "うなずく<end>": "고개를 끄덕인다<end>",
    "助けよう<end>": "도와주자<end>",
    "知ってる<end>": "알고 있어<end>",
    "知らない<end>": "몰라<end>",
    "人殺し！<end>": "살인자!<end>",
    "くそっ！<end>": "젠장!<end>",
    "その通り<end>": "맞아<end>",
    "いや違う<end>": "아니, 틀려<end>",
    "いいえ！<end>": "아니!<end>",
    "戻らない<end>": "돌아가지 않는다<end>",

    # Context-independent acknowledgements and pain/frustration interjections.
    "はい。<end>": "네.<end>",
    "はい！<end>": "네!<end>",
    "はい？<end>": "네?<end>",
    "ええ。<end>": "네.<end>",
    "いいえ<end>": "아니요<end>",
    "ちっ。<end>": "칫.<end>",
    "ちっ！<end>": "칫!<end>",
    "…ちっ<end>": "…칫<end>",
    "ちぇ。<end>": "쳇.<end>",
    "あ…。<end>": "아…<end>",
    "…あ。<end>": "…아.<end>",
    "うっ。<end>": "윽.<end>",
    "うっ！<end>": "윽!<end>",
    "うっ！！<end>": "윽!!<end>",
    "くっ！<end>": "큭!<end>",
    "くっ…<end>": "큭…<end>",
    "ぐっ！<end>": "큭!<end>",
    "ぐわっ！<end>": "크악!<end>",
    "痛っ！<end>": "아야!<end>",
    "はぁ…<end>": "하아…<end>",
    "うん…<end>": "응…<end>",
    "うん。<end>": "응.<end>",
    "うん！<end>": "응!<end>",
    "え…？<end>": "어…?<end>",
    "えっ？<end>": "어?<end>",
}


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("packet", type=Path)
    ap.add_argument("--out", type=Path)
    args = ap.parse_args()

    rows: list[str] = []
    with args.packet.open(encoding="utf-8") as f:
        for lineno, line in enumerate(f, 1):
            if not line.strip():
                continue
            obj = json.loads(line)
            ja = str(obj["japanese"])
            ko = EXACT.get(ja)
            if ko is None:
                continue
            rows.append(json.dumps(
                {"section": int(obj["section"]), "id": str(obj["id"]), "korean": ko},
                ensure_ascii=False,
                separators=(",", ":"),
            ) + "\n")

    payload = "".join(rows)
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(payload, encoding="utf-8")
    else:
        print(payload, end="")
    print(f"auto-trivial: {len(rows)} rows", file=__import__("sys").stderr)


if __name__ == "__main__":
    main()
