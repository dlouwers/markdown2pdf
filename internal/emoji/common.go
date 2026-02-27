package emoji

// CommonEmoji lists the top 100 most frequently used emoji that should be
// rendered as color PNG images from Twemoji. Emoji not in this list fall
// back to the Noto Emoji font (black & white).
//
// Coverage: Unicode 13.0+ most common emoji across messaging, documentation,
// and developer tools.
var CommonEmoji = map[rune]bool{
	// Smileys & Emotion (top picks)
	'😀': true, // grinning face
	'😁': true, // beaming face with smiling eyes
	'😂': true, // face with tears of joy
	'🤣': true, // rolling on the floor laughing
	'😃': true, // grinning face with big eyes
	'😄': true, // grinning face with smiling eyes
	'😅': true, // grinning face with sweat
	'😆': true, // grinning squinting face
	'😉': true, // winking face
	'😊': true, // smiling face with smiling eyes
	'😋': true, // face savoring food
	'😎': true, // smiling face with sunglasses
	'😍': true, // smiling face with heart-eyes
	'😘': true, // face blowing a kiss
	'🥰': true, // smiling face with hearts
	'😗': true, // kissing face
	'😙': true, // kissing face with smiling eyes
	'😚': true, // kissing face with closed eyes
	'🙂': true, // slightly smiling face
	'🤗': true, // hugging face
	'🤩': true, // star-struck
	'🤔': true, // thinking face
	'🤨': true, // face with raised eyebrow
	'😐': true, // neutral face
	'😑': true, // expressionless face
	'😶': true, // face without mouth
	'🙄': true, // face with rolling eyes
	'😏': true, // smirking face
	'😣': true, // persevering face
	'😥': true, // sad but relieved face
	'😮': true, // face with open mouth
	'🤐': true, // zipper-mouth face
	'😯': true, // hushed face
	'😪': true, // sleepy face
	'😫': true, // tired face
	'🥱': true, // yawning face
	'😴': true, // sleeping face
	'😌': true, // relieved face
	'😛': true, // face with tongue
	'😜': true, // winking face with tongue
	'😝': true, // squinting face with tongue
	'🤤': true, // drooling face
	'😒': true, // unamused face
	'😓': true, // downcast face with sweat
	'😔': true, // pensive face
	'😕': true, // confused face
	'🙃': true, // upside-down face
	'🤑': true, // money-mouth face
	'😲': true, // astonished face
	'☹': true, // frowning face
	'🙁': true, // slightly frowning face
	'😖': true, // confounded face
	'😞': true, // disappointed face
	'😟': true, // worried face
	'😤': true, // face with steam from nose
	'😢': true, // crying face
	'😭': true, // loudly crying face
	'😦': true, // frowning face with open mouth
	'😧': true, // anguished face
	'😨': true, // fearful face
	'😩': true, // weary face
	'🤯': true, // exploding head
	'😬': true, // grimacing face
	'😰': true, // anxious face with sweat
	'😱': true, // face screaming in fear
	'🥵': true, // hot face
	'🥶': true, // cold face
	'😳': true, // flushed face
	'🤪': true, // zany face
	'😵': true, // dizzy face
	'🥴': true, // woozy face
	'😠': true, // angry face
	'😡': true, // pouting face

	// Hand gestures
	'👍': true, // thumbs up
	'👎': true, // thumbs down
	'👌': true, // OK hand
	'✌': true, // victory hand
	'🤞': true, // crossed fingers
	'🤟': true, // love-you gesture
	'🤘': true, // sign of the horns
	'🤙': true, // call me hand
	'👈': true, // backhand index pointing left
	'👉': true, // backhand index pointing right
	'👆': true, // backhand index pointing up
	'👇': true, // backhand index pointing down
	'☝': true, // index pointing up
	'✋': true, // raised hand
	'🤚': true, // raised back of hand
	'🖐': true, // hand with fingers splayed
	'🖖': true, // vulcan salute
	'👋': true, // waving hand
	'🤝': true, // handshake
	'🙏': true, // folded hands
	'✍': true, // writing hand
	'💪': true, // flexed biceps
	'🦵': true, // leg
	'🦶': true, // foot
	'👏': true, // clapping hands

	// Objects & Symbols
	'❤': true, // red heart
	'🧡': true, // orange heart
	'💛': true, // yellow heart
	'💚': true, // green heart
	'💙': true, // blue heart
	'💜': true, // purple heart
	'🖤': true, // black heart
	'🤍': true, // white heart
	'🤎': true, // brown heart
	'💔': true, // broken heart
	'❣': true, // heart exclamation
	'💕': true, // two hearts
	'💞': true, // revolving hearts
	'💓': true, // beating heart
	'💗': true, // growing heart
	'💖': true, // sparkling heart
	'💘': true, // heart with arrow
	'💝': true, // heart with ribbon
	'💟': true, // heart decoration
	'☮': true, // peace symbol
	'✝': true, // latin cross
	'☪': true, // star and crescent
	'🕉': true, // om
	'☸': true, // wheel of dharma
	'✡': true, // star of david
	'🔯': true, // dotted six-pointed star
	'🕎': true, // menorah
	'☯': true, // yin yang
	'☦': true, // orthodox cross
	'🛐': true, // place of worship
	'⛎': true, // ophiuchus
	'♈': true, // aries
	'♉': true, // taurus
	'♊': true, // gemini
	'♋': true, // cancer
	'♌': true, // leo
	'♍': true, // virgo
	'♎': true, // libra
	'♏': true, // scorpio
	'♐': true, // sagittarius
	'♑': true, // capricorn
	'♒': true, // aquarius
	'♓': true, // pisces

	// Activity & Objects
	'⚽': true, // soccer ball
	'🏀': true, // basketball
	'🏈': true, // american football
	'⚾': true, // baseball
	'🥎': true, // softball
	'🎾': true, // tennis
	'🏐': true, // volleyball
	'🏉': true, // rugby football
	'🥏': true, // flying disc
	'🎱': true, // pool 8 ball
	'🪀': true, // yo-yo
	'🏓': true, // ping pong
	'🏸': true, // badminton
	'🏒': true, // ice hockey
	'🏑': true, // field hockey
	'🥍': true, // lacrosse
	'🏏': true, // cricket game
	'🥅': true, // goal net
	'⛳': true, // flag in hole
	'🪁': true, // kite
	'🏹': true, // bow and arrow
	'🎣': true, // fishing pole
	'🤿': true, // diving mask
	'🥊': true, // boxing glove
	'🥋': true, // martial arts uniform
	'🎽': true, // running shirt
	'🛹': true, // skateboard
	'🛼': true, // roller skate
	'🛷': true, // sled
	'⛸': true, // ice skate
	'🥌': true, // curling stone
	'🎿': true, // skis
	'⛷': true, // skier
	'🏂': true, // snowboarder
	'🪂': true, // parachute
	'🏋': true, // person lifting weights
	'🤼': true, // people wrestling
	'🤸': true, // person cartwheeling
	'🤺': true, // person fencing
	'⛹': true, // person bouncing ball
	'🤾': true, // person playing handball
	'🏌': true, // person golfing
	'🏇': true, // horse racing
	'🧘': true, // person in lotus position
	'🏄': true, // person surfing
	'🏊': true, // person swimming
	'🤽': true, // person playing water polo
	'🚣': true, // person rowing boat
	'🧗': true, // person climbing
	'🚵': true, // person mountain biking
	'🚴': true, // person biking
	'🏆': true, // trophy
	'🥇': true, // 1st place medal
	'🥈': true, // 2nd place medal
	'🥉': true, // 3rd place medal
	'🏅': true, // sports medal
	'🎖': true, // military medal
	'🏵': true, // rosette
	'🎗': true, // reminder ribbon
	'🎫': true, // ticket
	'🎟': true, // admission tickets
	'🎪': true, // circus tent
	'🤹': true, // person juggling
	'🎭': true, // performing arts
	'🩰': true, // ballet shoes
	'🎨': true, // artist palette
	'🎬': true, // clapper board
	'🎤': true, // microphone
	'🎧': true, // headphone
	'🎼': true, // musical score
	'🎹': true, // musical keyboard
	'🥁': true, // drum
	'🪘': true, // long drum
	'🎷': true, // saxophone
	'🎺': true, // trumpet
	'🪗': true, // accordion
	'🎸': true, // guitar
	'🪕': true, // banjo
	'🎻': true, // violin
	'🪈': true, // flute
	'🎲': true, // game die
	'♟': true, // chess pawn
	'🎯': true, // direct hit
	'🎳': true, // bowling
	'🎮': true, // video game
	'🎰': true, // slot machine
	'🧩': true, // puzzle piece

	// Common status/indicator emoji
	'✅': true, // check mark button
	'❌': true, // cross mark
	'⭐': true, // star
	'⚠': true, // warning
	'🔥': true, // fire
	'💯': true, // hundred points
	'✨': true, // sparkles
	'🚀': true, // rocket
	'💡': true, // light bulb
	'🎉': true, // party popper
	'🎊': true, // confetti ball
	'🎈': true, // balloon
	'🎁': true, // wrapped gift
	// Trophy already defined above at line 209
	'🔴': true, // red circle
	'🟠': true, // orange circle
	'🟡': true, // yellow circle
	'🟢': true, // green circle
	'🔵': true, // blue circle
	'🟣': true, // purple circle
	'⚫': true, // black circle
	'⚪': true, // white circle
	'🟤': true, // brown circle
	'🔸': true, // small orange diamond
	'🔹': true, // small blue diamond
	'🔶': true, // large orange diamond
	'🔷': true, // large blue diamond
	'🔺': true, // red triangle pointed up
	'🔻': true, // red triangle pointed down
	'💠': true, // diamond with a dot
	'🔘': true, // radio button
	'🔳': true, // white square button
	'🔲': true, // black square button
}

// IsCommonEmoji checks if a rune is in the common emoji set that should
// be rendered as a color PNG.
func IsCommonEmoji(r rune) bool {
	return CommonEmoji[r]
}
