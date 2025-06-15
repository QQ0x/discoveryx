package text

import (
	"golang.org/x/text/language"
)

type TextID int

const (
	TextIDStart TextID = iota
	TextIDStartLunker
	TextIDStartTouch
	TextIDOpening
	TextIDEnding
	TextIDEndingScore1
	TextIDEndingScore2
	TextIDEndingScore3
	TextIDSecretCommand
	TextIDSecretClear
	TextIDItemPowerUp
	TextIDItemFuji
	TextIDItemBushi
	TextIDItemApple
	TextIDItemV
	TextIDItemTaka
	TextIDItemShoulder
	TextIDItemDagger
	TextIDItemKatakata
	TextIDItemNasu
	TextIDItemBonus
	TextIDItemNurse
	TextIDItemNazuna
	TextIDItemGameHell
	TextIDItemGundam
	TextIDItemPoed
	TextIDItemMilestone
	TextIDItem1Yen
	TextIDItemTriangle
	TextIDItemOmega
	TextIDItemLife
)

var texts = map[language.Tag]map[TextID]string{
	language.Japanese: {
		TextIDStart:       "すぺーす　たたく　はじまる！",
		TextIDStartLunker: "らんかー　もーど　はじまる！",
		TextIDStartTouch:  "がめん　たっち　はじまる！",
		TextIDOpening: `めが　さめたら
<red>いのしし</red>に　なっていた。



おちつけ　おれ。
＊＊　ゆめの　なかにいる　＊＊



ゆめの　ちとゅう
どるいどの　よげんしゃが
わたしに　つげるのだった。



「<red>さんしゅの　じんぎ</red>を　みつけだせるのは
この　ゆめの　しゅじんこう…
そう　そなただけなのじゃ！」



いかなる　しょうがいをも　のりこえ
はつゆめ　おやくそくの
<red>さんしゅの　じんぎ</red>を　さがしだすこと…
それが　わたしの
はたすべき　<red>しめい</red>なのだ！



わたしの　いのししとしての　<red>ちが　さわぐ</red>！`,
		TextIDEnding: `あつめた、<red>じんぎ</red>たちが　かがやきだす！



うおーっ！

わたしは　さけびごえを　あげ
ひかりの　なかへ…

ほっぷ　すてっぷ　じゃんぷ
かーるいす！



やがて　ひとつの　いんしは
その　いしを　もとの　ばしょへ
かいきさせ
きおくの しんえんに　きざまれた
きげんの　いしきを
おもい　おこさせる　だろう…





2007　しんねん　はじまる！
<red>あけましておめでとう！</red>





<red>――</red>　ここから　くれじっと　<red>――</red>


いのべーじょん2007
＊＊ くれじっとの なかに いる ＊＊



すーぱー　ざつよう　にんげん
おめが　／　（　゜ワ゜）ノ

おんがく　にんげん
どん

ふいーるど　まっぷ　にんげん
おめが　／　（　゜ワ゜）ノ
げっく
ずい
３５１

たいとる　めいめい　にんげん
わんきち

てすとぷれい　にんげん
げーむへる2000
げっく
３５１
ずぃ
あかじゃ

すべさる　さんくす　にんげん
げーむ　せいさく　ぎじゅつ　いた
どにちまでに　げーむを　つくる　すれ

HTML5　いしょく
はねだ

ごー　げんご　いしょく
ほしはじめ


ぷろどぅーすど　ばい
おめが



<red>――</red>　くれじっと　ここまで　<red>――</red>

えんどう　おふろに　はいる`,
		TextIDEndingScore1: "せいせき　はぴょう",
		TextIDEndingScore2: "かくとく　あいてむ",
		TextIDEndingScore3: "くりあ　たいむ",
		TextIDSecretCommand: `たいとるで

ひだり ひだり ひだり
みぎ　みぎ　みぎ
ひだり　みぎ`,
		TextIDSecretClear: `おめでとう。

あなたは
ぎじゅつ　さる
です。`,
		TextIDItemPowerUp: `「みずぐすり」を　てに　いれた。
「てーれってれー」
「じゃんぷりょくが　あっぷ。
<red>くうちゅう　じゃんぷ</red>が　１かい
くわわる　くわわる！`,
		TextIDItemFuji: `「ふじ」を　てに　いれた。
じんぎの　ひとつ。
ふんかしない　かざん。
きゅうかざん　って
いうんだって。`,
		TextIDItemBushi: `「ぶし」を　てに　いれた。
じゃぱにーず　ないと。
あめりか　だいとうりょうも　ぶしどー。
さむらーい　さむらーい　ぶしどー。`,
		TextIDItemApple: `「ふじりんご」を　てに　いれた。
あかい　かじつ。
あーかーい　りんごーにー
くちびーる　よーせーてー。`,
		TextIDItemV: `「ぶい」を　てに　いれた。
たたかいの　さけび！
「しんかとは　ひとと　げーむの
がったいだ！」`,
		TextIDItemTaka: `「たか」を　てに　いれた。
じんぎの　ひとつ。
そらの　はんたー。
こだいあすてかでは
かみの　つかい　なんだ。`,
		TextIDItemShoulder: `「かた」を　てに　いれた。
どうたいの　うえ、
うでの　つけね。
かたが　あかいと
じかんに　おくれる。`,
		TextIDItemDagger: `「だがー」を　てに　いれた。
みじかい　けん
ひだりてたての　かわりに。
こがたなの　かわりに。
ぼうけんの　おともに　どうぞ。`,
		TextIDItemKatakata: `「かたかた」を　てに　いれた。
かたかた…
「これは　まるたーがいすと…」
「ちがう！ぷらずま　だ！！」`,
		TextIDItemNasu: `「なす」を　てに　いれた。
じんぎの　ひとつ。
むらさきに　かがやく
やさいの　おうさま
でも　あまり　たべたきが　しない。`,
		TextIDItemBonus: `「ぼうなす」を　てに　いれた。
あ　ぼうなす！
ふぇいたりてぃ　ぼうなす！
ぱしふぃすと　ぼうなす！
でぃす　いず　ざ　ぼうなす！`,
		TextIDItemNurse: `「なーす」を　てに いれた。
「かんごふでは　ない！
『かんごし』と　よべ！
この　ペいしえんと　どもめ！」`,
		TextIDItemNazuna: `「なずな」を　てに　いれた。
べつめい、ぺんぺんぐさ。
なずなが　とおったあとには
ぺんぺんぐさすら　のこらないという`,
		TextIDItemGameHell: `「げーむへる」を　てに　いれた。
ようこそ。
げーむ　せいさくしゃと
その　しゅうへんの　ための
こみゅにていへ。`,
		TextIDItemGundam: `「じっしゃがんだむ」を　てに　いれた。
「そうさは　かんたんだ。
せんとう　こんぴゅーたの
すいっちを　いれるだけで　いい。」`,
		TextIDItemPoed: `「ほえど」を　てに　いれた。
「ふん。
おもしろくなって
きやがったぜ！」`,
		TextIDItemMilestone: `「まいるまーく」を　てに　いれた。
さいしんさく「からす」
ぜっさん　かどうちゅう　でしゅー`,
		TextIDItem1Yen: `「いちえんさつ」を　てに　いれた。
「くらくて　よくみえないわ」
「ほうら　あかるいだろう」
「『くらくて　よくみえないわ』と
かいてある」`,
		TextIDItemTriangle: `「とらいあんぐる」を　てに　いれた。
すべって　ころんで　おおいたけん。
しゃちょうは　いま
どうしているのか…`,
		TextIDItemOmega: `「おめがの　くんしょう」を　てに　いれた。
こんな　げーむに
まじに　なって
どうも　ありがとう`,
		TextIDItemLife: `「はーとの　うつわ」を　てに　いれた。
でれででーん！
<red>らいふ</red>の　<red>じょうげん</red>を
１ふやしてあげる
ああ、なんて　たくましいの…`,
	},
	language.English: {
		TextIDStart:       "PRESS SPACE BEGIN!",
		TextIDStartLunker: "LUNKER MODE BEGIN!",
		TextIDStartTouch:  "TOUCH SCREEN BEGIN!",
		TextIDOpening: `when I wake up
I become <red>Wild bore</red>
calm down me
** stay in dream **





in a dream
a Prophet Druid man
He telling me
can finding <red>THE Three "GOT ICONs"</red>
this dream's HERO
you. just only YOU!





beat any traps
Telling of
"The first dream of the year of Japan"
explor <red>"Imperial Regalia of Japan"</red>
This is the <red>FATE</red> of me





My WILD BORE's Heart get <red>excited</red>.`,
		TextIDEnding: `shining <red>"Imperial Regalia of Japan"</red>
OOOMPF!



Hop Step Jump
carllouis!



Before long A factor was
her Intention had calling to original place
Recurrence...
Memoried to His Heart
Recall it makes Mind of The Creation


2007 spring has come!


<red>a Happy New Year!</red>





INNOVATION 2007



Super Trivial Works
O-MEGA



composer man
dong



Field Maps Design
O-MEGA
Gekku
Zy
351



Title naming
wang-zhi



Test Playing man
Game-Hell 2000
Gek
351
Zi
Akaja



EngRish translation
Eki
Dong
O-MEGA



Special Thanks man
Game Dev. BBS users
www.2ch.net Thread "Game Deveropment"



Porting to HTML5
Haneda



Porting to Go
Hajime Hoshi



Produced by
O-MEGA



EVZO-END get bath`,
		TextIDEndingScore1: "Results",
		TextIDEndingScore2: "You Got ICONS",
		TextIDEndingScore3: "Clear Time",
		TextIDSecretCommand: `at Title Screen

L L L
R R R
L R`,
		TextIDSecretClear: `Congratulations!

You are
the
technical monkey.`,
		TextIDItemPowerUp: `You Grab "Water Medicine"
TA-RA-TERAH!!
JUMP UP
you got "M. Jordan's AIR NIKE"
increase INCREASES!`,
		TextIDItemFuji: `You Grab a "FUJIYAMA"
*An Imperial Regalia*
Throw the One Ring Frodo!`,
		TextIDItemBushi: `You Feel "BUSHI-DO"
Japanese SAMURAI Spilit
USMC, Marines have BUSHI-DO
SAMURAI SAMURAI BUSHI-DO`,
		TextIDItemApple: `You Grab Big Apple(fuji-ringo)
Red Sweety
Let it be lalala`,
		TextIDItemV: `You Grab "V of Victory"
war cry!
Me Humanbeing with a VIDEO GAME
PILDER ON!`,
		TextIDItemTaka: `You catch a Hawk(Taka).
*an Imperial Regalia*
The President of Sky.
In ancient Aztec
A messenger of God`,
		TextIDItemShoulder: `You got a RED-SHOULDER(Kata)
upper bodie The root of Arms
Red-Shoulder`,
		TextIDItemDagger: `You Got a Dagger
Short sword
the only friend
on your quest`,
		TextIDItemKatakata: `You Grab "Katakata"
strange sound
RATTLE-RATTLE...
Poltergeist!?
NO! ask to Mulder!`,
		TextIDItemNasu: `You Got Eggplant (Nasu)
*An Imperial Regalia*
Purple Shining
King of Vegetable
But I've never satisfied with this`,
		TextIDItemBonus: `You Got "Bonus"
A Bonus!
Fatality Bonus!
Pacifist Bonus!
This is the Bonus!`,
		TextIDItemNurse: `You Grab "Nurse"
I'm not an angel in white
I'm a goddess of hell
You Patients!`,
		TextIDItemNazuna: `You got "Nazuna"
It is called
"Shepherd's-purse"
in English
Symbol of doom`,
		TextIDItemGameHell: `You Grab “GameHel12000"
Helcome.
We are an community of
Japanese indie game developers.`,
		TextIDItemGundam: `You Grab "GUNDAM 0079"
Control is easy
All you have to do is
switching on a battle computer`,
		TextIDItemPoed: `You Grab "Po'ed"
Ha!
It's getting more interesting!`,
		TextIDItemMilestone: `You got "Milestone Mark"
All your wii are belong to
our doomsday device KAROUS.`,
		TextIDItem1Yen: `You Got 1 Yen.
In ancient Japan
It was used as a
substitute for light`,
		TextIDItemTriangle: `You Got "Triangle Service"
Slip Trip Oita-Pref.
Where is the Boss?`,
		TextIDItemOmega: `You Got "O-mega Medal"
Thank you for playing
such a masochistic game.`,
		TextIDItemLife: `You Got a Heart Container
DeDeDe-Deng!!
Increase your MAX LIFE.
REAL MEN!!`,
	},
}

func Get(lang language.Tag, id TextID) string {
	return texts[lang][id]
}

func Languages() []language.Tag {
	// TODO: Sort
	langs := []language.Tag{}
	for lang := range texts {
		langs = append(langs, lang)
	}
	return langs
}
