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

    # Stable standalone NPC/job labels.  These deliberately exclude dialogue
    # fragments whose Korean register depends on speaker/context.
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
    "宿泊客<end>": "투숙객<end>",
    "常連客<end>": "단골손님<end>",
    "見物客<end>": "구경꾼<end>",
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
