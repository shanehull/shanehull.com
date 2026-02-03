package handlers

type Quote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

var quotes = []Quote{
	{
		Text:   "With all things being equal, the simplest explanation tends to be the right one.",
		Author: "William of Ockham",
	},
	{
		Text:   "The first principle is that you must not fool yourself - and you are the easiest person to fool.",
		Author: "Richard P. Feynman",
	},
	{
		Text:   "To decide upon the answer is not scientific. In order to make progress, one must leave the door to the unknown ajar–ajar only.",
		Author: "Richard P. Feynman",
	},
	{
		Text:   "There will be no order, only chaos.",
		Author: "Sol Robeson",
	},
	{
		Text:   "There are two kinds of forecasters: those who don't know, and those who don't know they don't know.",
		Author: "John Kenneth Galbraith",
	},
	{
		Text:   "It is much more sound to take risks you can measure than to measure the risks you are taking.",
		Author: "Nassim Nicholas Taleb",
	},
	{
		Text:   "Mother nature does not like overspecialisation, as it limits evolution and weakens the animals.",
		Author: "Nassim Nicholas Taleb",
	},
	{
		Text:   "You don't become completely free just by avoiding to be a slave; you also need to avoid becoming a master.",
		Author: "Nassim Nicholas Taleb",
	},
	{
		Text:   "The characteristic feature of the loser is to bemoan, in general terms, mankind's flaws, biases, contradictions, and irrationality – without exploiting them for fun and profit.",
		Author: "Nassim Nicholas Taleb",
	},
	{
		Text:   "The idea is that flowing water never goes stale, so just keep on flowing.",
		Author: "Lee Jun-fan",
	},
	{
		Text:   "Absorb what is useful, discard what is useless and add what is specifically your own.",
		Author: "Lee Jun-fan",
	},
	{
		Text:   "The idea that the future is unpredictable is undermined every day by the ease with which the past is explained.",
		Author: "Daniel Kahneman",
	},
	{
		Text:   "No man ever steps in the same river twice, for it's not the same river and he's not the same man.",
		Author: "Heraclitus",
	},
	{
		Text:   "I am the wisest man alive, for I know one thing, and that is that I know nothing.",
		Author: "Socrates",
	},
	{
		Text:   "If the only tool you have is a hammer, you tend to see every problem as a nail.",
		Author: "Abraham Maslow",
	},
	{
		Text:   "I think that a life properly lived is just learn, learn, learn all the time.",
		Author: "Charles T. Munger",
	},
	{
		Text:   "Show me the incentive and I'll show you the outcome.",
		Author: "Charles T. Munger",
	},
	{
		Text:   "History never repeats itself. Man always does.",
		Author: "M. de Voltaire",
	},
	{
		Text:   "The impediment to action advances action. What stands in the way becomes the way.",
		Author: "Marcus Aurelius",
	},
	{
		Text:   "Think of yourself as dead. You have lived your life. Now take what's left and live it properly.",
		Author: "Marcus Aurelius",
	},
	{
		Text:   "Do not seek to follow in the footsteps of the wise; seek what they sought.",
		Author: "Matsuo Basho",
	},
	{
		Text:   "Risk is what's left over after you think you've thought of everything.",
		Author: "Carl Richards",
	},
	{
		Text:   "To know what you know and what you do not know, that is true knowledge.",
		Author: "Kongzi",
	},
	{
		Text:   "The world is ruled by letting things take their course. It cannot be ruled by interfering.",
		Author: "Laozi",
	},
	{
		Text:   "Force is followed by loss of strength.",
		Author: "Laozi",
	},
	{
		Text:   "Everything should be made as simple as possible, but no simpler.",
		Author: "Not Einstein",
	},
	{
		Text:   "To think is easy. To act is hard. But the hardest thing in the world is to act in accordance with your thinking.",
		Author: "Johann Wolfgang von Goethe",
	},
	{
		Text:   "The most dangerous government is the one that's closest to you.",
		Author: "Rick Rule",
	},
	{
		Text:   "Never ascribe to malice that which can be adequately explained by stupidity.",
		Author: "Robert J. Hanlon",
	},
	{
		Text:   "In theory, there is no difference between theory and practice. But in practice, there is.",
		Author: "Yogi Berra",
	},
	{
		Text:   "The geometrician teaches me how I may avoid losing any fraction of my estates, but what I really want to learn is how to lose the lot and still keep smiling.",
		Author: "Lucius Annaeus Seneca",
	},
	{
		Text:   "It is better to be vaguely right than exactly wrong.",
		Author: "Carveth Read",
	},
	{
		Text:   "All of humanity's problems stem from man's inability to sit quietly in a room alone.",
		Author: "Blaise Pascal",
	},
	{
		Text:   "Man selects only for his own good: Nature only for that of the being which she tends.",
		Author: "Charles Darwin",
	},
}
