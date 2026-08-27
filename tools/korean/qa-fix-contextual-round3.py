#!/usr/bin/env python3
from __future__ import annotations
import json,re,tomllib
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
KDIR=ROOT/'translations'/'korean'/'messages'
FRE=re.compile(r'^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$')
HDR=re.compile(r'^\["(?P<id>\d+)"\]$')
KORE=re.compile(r'^korean = (?P<v>".*")$')
WS=re.compile(r'[\s\u3000]+')
def norm(s:str)->str:return WS.sub('',s.replace('<line-break>',''))
CAN={
 norm('しかし、さすがは破壊神の円卓騎士ですよね。猫の姿になってなお人語を操るんですから。<end>'):'그래도 역시 파괴신의 원탁기사답네요. 고양이 모습이 되고도 사람 말을 하니까요.<end>',
 norm('だから、錬剛石がとれる山を見つけたら、何度か足を運んでみるのがいいだろうな。<end>'):'그러니 연강석이 나는 산을 찾으면 몇 번쯤 찾아가 보는 게 좋겠지.<end>',
 norm('なんとか、たどり着けたな。役目ご苦労だった。約束の報酬だ。受け取り給え。<end>'):'어떻게든 도착했군. 임무 수고했다. 약속한 보수다. 받아라.<end>',
 norm('ま、こんな感じで、これは遠くの人を呼び寄せたり…。さて、それでは、もう一度。<end>'):'뭐, 이런 식으로 멀리 있는 사람을 불러오기도 하고…. 자, 그럼 다시 한번.<end>',
 norm('ディンガル帝国は、アンギルダンという優秀な将軍を失いながらも、その勢いを損なうことなく大陸に覇をとなえようとしていた…。<end>'):'딩갈 제국은 앙길단이라는 뛰어난 장군을 잃었음에도 기세를 잃지 않고 대륙에 패권을 떨치려 하고 있었다….<end>',
 norm('ホント…昔から人間って好きになれなかったわ…。弱いくせに、生意気でいつでも裏をかくことを考えてて油断ならなくて…。<end>'):'정말… 옛날부터 인간은 좋아할 수가 없었어…. 약하면서 건방지고, 늘 뒤통수칠 생각이나 하고, 방심할 수가 없어서….<end>',
 norm('世界を巡って、人に出会い、仲間を作ってください。その中で世界は様々な顔を見せてくれるでしょう。自由な旅を！<end>'):'세계를 돌아다니며 사람을 만나고 동료를 만드세요. 그 속에서 세계는 여러 얼굴을 보여 줄 겁니다. 자유로운 여행을!<end>',
 norm('忘れないように覚えておかなくちゃ。<end>'):'잊지 않도록 기억해 둬야겠어.<end>',
 norm('破壊神ウルグの円卓騎士と呼ばれる強力な魔人の１人です。妖艶な美女で、残酷で気まぐれです。闇の神器をひとつ手に入れるためにミイスを焼き払ってしまいました。<end>'):'파괴신 울그의 원탁기사라 불리는 강력한 마인 중 한 명입니다. 요염한 미녀이며 잔혹하고 변덕스럽습니다. 어둠의 신기 하나를 얻기 위해 미이스를 불태웠습니다.<end>',
 norm('言葉をしゃべるゴブリンなんて今まで聞いたことがない。な、怪しいと思うだろ？<end>'):'말하는 고블린이라니 지금까지 들어 본 적도 없어. 수상하다고 생각하지?<end>',
 norm('あっ、そうだ、わたしの仲間を紹介します！<end>'):'아, 맞다. 제 동료를 소개할게요!<end>',
 norm('あなたが将来、旅で人と出会い、仲間を増やしたとき、一緒に旅する仲間を変更をしたいと思ったときここに来てください。<end>'):'앞으로 여행 중 사람을 만나 동료가 늘고, 함께 여행하는 동료를 바꾸고 싶어지면 여기로 오세요.<end>',
 norm('あのときは、<value:$28>を守るので必死だったからね。親というものはそんなものだよ。<end>'):'그때는 <value:$28>을 지키느라 필사적이었으니까. 부모란 그런 거야.<end>',
 norm('い？それが何かだって？そんなものは人に聞くもんじゃないんだよ。若いんだから、怠けてねぇで自分で見つけなよ。<end>'):'응? 그게 뭐냐고? 그런 걸 남에게 묻는 게 아니야. 젊은데 게으름 피우지 말고 스스로 찾아.<end>',
 norm('この方のことならば少し聞いている。信頼に足る方だ。<end>'):'이분 이야기는 조금 들었습니다. 신뢰할 만한 분입니다.<end>',
 norm('さすがに鋭いですね。その通り。その聖杯こそ禁断の聖杯です。ですが、それを盗み出させたのは魔人アーギルシャイアではない別の魔人なのです。<end>'):'역시 날카롭군요. 맞습니다. 그 성배가 바로 금단의 성배입니다. 하지만 그걸 훔치게 한 건 마인 아르길샤이어가 아니라 다른 마인입니다.<end>',
}
PARTICLE_FIXES={
 'バロル':(('발로르은','발로르는'),('발로르과','발로르와')),
 'ゾフォル':(('조포르이','조포르가'),),
}
seen=set(); changed=files=0
for p in sorted(KDIR.glob('msgsec*.toml')):
 if not FRE.match(p.name):continue
 with p.open('rb') as f:data=tomllib.load(f)
 lines=p.read_text(encoding='utf-8').splitlines(); cur=None; dirty=False
 for i,line in enumerate(lines):
  m=HDR.match(line)
  if m:cur=m.group('id');continue
  if cur is None or not KORE.match(line):continue
  try:n=int(cur)
  except:cur=None;continue
  if n in seen:cur=None;continue
  seen.add(n); rec=data.get(cur); cur=None
  if not isinstance(rec,dict):continue
  ja=rec.get('japanese');ko=rec.get('korean')
  if not isinstance(ja,str) or not isinstance(ko,str):continue
  new=CAN.get(norm(ja),ko)
  for ja_term,repls in PARTICLE_FIXES.items():
   if ja_term not in ja:continue
   for old,replacement in repls:new=new.replace(old,replacement)
  if new!=ko:
   lines[i]='korean = '+json.dumps(new,ensure_ascii=False);dirty=True;changed+=1
 if dirty:
  p.write_text('\n'.join(lines)+'\n',encoding='utf-8');files+=1
print(f'Contextual round 3: {changed} records across {files} files')
