package polygon

import "strings"

var tagMapping = map[string][]string{
	"2-sat":                     {"4qpkrclrfl7rv5lic8djr3lldk"}, // 2-SAT
	"2d array":                  {"pjjft5joql5j95u7radbchs51g"}, // For Beginners
	"ad-hoc":                    {},
	"adhoc":                     {},
	"arrays":                    {"pjjft5joql5j95u7radbchs51g"}, // For Beginners
	"backtracking":              {"amb5r8c4bt1395uneribtcnces"}, // Depth-first Search
	"bacs_review":               {},
	"beginner":                  {"pjjft5joql5j95u7radbchs51g"},                               // For Beginners
	"bfs":                       {"c79dhqpr712uv2feapbaodn6ds"},                               // Breadth-first Search
	"binary search":             {"3hr591p5lh7a9c5k9bg8kpvctg"},                               // Binary Search
	"binsearch":                 {"3hr591p5lh7a9c5k9bg8kpvctg"},                               // Binary Search
	"bitmask-dp":                {"ab0l1n5rsl3618ntv9ode0qn5k", "agd25lpqt52m565ljb65l0mqg0"}, // Bitmasks, Dynamic programming
	"bitmasks":                  {"ab0l1n5rsl3618ntv9ode0qn5k"},                               // Bitmasks
	"bitwise operation":         {"mnvqr2ggu10jl7r8kpmdsusqlk"},                               // Bitwise Operations
	"bracket sequences":         {},
	"brute force":               {"5cl1ftokid1bn751ql21o0vdbs"}, // Brute force
	"chinese remainder theorem": {"5nletoaur90j97jm9ac29rtkts"}, // Chinese remainder theorem
	"codework":                  {},
	"combinatorics":             {"66aoi354mt23da5mrt4npes0o0"}, // Combinatorics
	"constructive":              {"7s78regvmt1ata79gr1ndlu67o"}, // Constructive Algorithms
	"constructive algorithms":   {"7s78regvmt1ata79gr1ndlu67o"}, // Constructive Algorithms
	"convex hull":               {"b03jtl2ah5371f5msag4pspv0g"}, // Convex Hull
	"data structures":           {"8asv7g7jbl7hjclnc5ehiiamcs"}, // Data Structures
	"dejkstra":                  {"httb8civtl0u74jm2e143pm5ok"}, // Dijkstra algorithm
	"deque":                     {},
	"dfs":                       {"amb5r8c4bt1395uneribtcnces"}, // Depth-first Search
	"dfs and similar":           {"amb5r8c4bt1395uneribtcnces"}, // Depth-first Search
	"dijkstra":                  {"httb8civtl0u74jm2e143pm5ok"}, // Dijkstra algorithm
	"disjoint set union":        {"eo9vh68hjd1kbf445f8aoi8rlc"}, // Disjoint set union
	"disjoint sets":             {"eo9vh68hjd1kbf445f8aoi8rlc"}, // Disjoint set union
	"div3":                      {"pjjft5joql5j95u7radbchs51g"}, // For Beginners
	"divide and conquer":        {"ad9840pj2d0rl10o0vja3c7q6g"}, // Divide and Conquer
	"djikstra":                  {"httb8civtl0u74jm2e143pm5ok"}, // Dijkstra algorithm
	"dp":                        {"agd25lpqt52m565ljb65l0mqg0"}, // Dynamic programming
	"dp optimization":           {"l7ehngruct3479vqegsnu8vng8"}, // Dynamic Programming Optimization
	"dsu":                       {"eo9vh68hjd1kbf445f8aoi8rlc"}, // Disjoint set union
	"dynamic programming":       {"agd25lpqt52m565ljb65l0mqg0"}, // Dynamic programming
	"easy":                      {"pjjft5joql5j95u7radbchs51g"}, // For Beginners
	"example":                   {},
	"expression parsing":        {"lqcb6ciath3crca477lrib36oo"}, // Expression parsing
	"factorization":             {"s6akn539f93sjfndpv3tm8j4io"}, // Prime factorization
	"fenwick tree":              {"drqgu3n5k10ep27aguknkvhbsk"}, // Fenwick Tree
	"fft":                       {"uoicsgaimp4f71sputplg5rd48"}, // Fast Fourier transform
	"flows":                     {"m2ouonsldt4cdd6mnuu5pm7kq8"}, // Graph network flows
	"for":                       {"pjjft5joql5j95u7radbchs51g"}, // For Beginners
	"formula":                   {},
	"game theory":               {"liijhm523122177op14gmgjd18"}, // Games
	"games":                     {"liijhm523122177op14gmgjd18"}, // Games
	"geometry":                  {"mn2buv28bp02v88uj715svnoro"}, // Geometry
	"graph":                     {"msh9q06gah6hpds3m0dceu3ff8"}, // Graphs
	"graph matchings":           {"mnki3h2qo51l91og62es1spa54"}, // Graph matchings
	"graph theory":              {"msh9q06gah6hpds3m0dceu3ff8"}, // Graphs
	"graphs":                    {"msh9q06gah6hpds3m0dceu3ff8"}, // Graphs
	"greedy":                    {"n0b0meiu7p51tekkqefeafqat0"}, // Greedy algorithms
	"hashing":                   {"n4irjrf3ot0rbdit566sbjrbio"}, // Hashing
	"if":                        {},
	"implementation":            {"nivqcdt8d93tff7rtkk7lu9ur8"}, // Implementation
	"interactive":               {"nlp1qosu1h7jj8k8t9dn2131rg"}, // Interactive
	"java":                      {},
	"joke":                      {},
	"lksh":                      {},
	"math":                      {"o44qcs7mvt6nj6k5qcliev933g"}, // Math
	"maths":                     {"o44qcs7mvt6nj6k5qcliev933g"}, // Math
	"matrices":                  {"onl6ffbeq56bpaskrv54o6qdlc"}, // Matrix
	"matrix":                    {"onl6ffbeq56bpaskrv54o6qdlc"}, // Matrix
	"matrix exponentiation":     {"onl6ffbeq56bpaskrv54o6qdlc"}, // Matrix
	"maxflow":                   {"m2ouonsldt4cdd6mnuu5pm7kq8"}, // Graph network flows
	"meet-in-the-middle":        {"34iiosa5s141r4msjhctjb6g74"}, // Meet-in-the-middle
	"number theory":             {"p1so0jh9k96m3f8t5u70l8nphc"}, // Number theory
	"optimization":              {"k0t2kb3p1d2r1clduv3nhlsnic"}, // Optimization
	"prefix sums":               {"h2pti09sm104fdee212ir6i4fs"}, // Prefix Sums
	"prefix-function":           {"0tdg3jl6857bn83t3t70mj0fkg"}, // Prefix Function
	"prime factorization":       {"s6akn539f93sjfndpv3tm8j4io"}, // Prime factorization
	"priority queue":            {"ojlb9b433d41f2lmiab34b98qc"}, // Priority Queue
	"probabilities":             {"pdoel1o5e936ve124idn9ar4dc"}, // Probabilities
	"queries":                   {"vsgktagv113fldeso88u2q9k38"}, // Queries
	"queue":                     {},
	"randomized-algorithms":     {"7snfna87rp69bbl5fln6i8bvqc"}, // Randomized algorithms
	"range queries":             {"uulanemp6l6lp9h5ja68k9u4pg"}, // Range queries
	"realization":               {"nivqcdt8d93tff7rtkk7lu9ur8"}, // Implementation
	"recursion":                 {},
	"rmq":                       {"uulanemp6l6lp9h5ja68k9u4pg"},                               // Range queries
	"scanline":                  {"f73kkr3fo932d0a0orkg8m4tb4"},                               // Scanline / Sweep Line
	"schedules":                 {"psh4svain501l2nbobe0njvmm0"},                               // Schedules
	"segment tree":              {"3arh7cff3t58l4ur4u1754iomg"},                               // Segment Tree
	"shortest paths":            {"q44u43ajtp7eb9bhn6e4h6mf1s"},                               // Shortest paths
	"sieve of eratosthenes":     {"p1so0jh9k96m3f8t5u70l8nphc"},                               // Number theory
	"simple":                    {"pjjft5joql5j95u7radbchs51g"},                               // For Beginners
	"simple math":               {"o44qcs7mvt6nj6k5qcliev933g", "pjjft5joql5j95u7radbchs51g"}, // Math, For Beginners
	"sorting":                   {"q5g5r2to9t3h11e9g7gsoeiius"},                               // Sorting
	"sortings":                  {"q5g5r2to9t3h11e9g7gsoeiius"},                               // Sorting
	"sqrt":                      {"u6atdab8ih6vv59u1768hcf85c"},                               // SQRT Decomposition
	"sqrt-decomposition":        {"u6atdab8ih6vv59u1768hcf85c"},                               // SQRT Decomposition
	"stack":                     {"8glqsj58uh5a17b171itjedcog"},                               // Stacks
	"stacks":                    {"8glqsj58uh5a17b171itjedcog"},                               // Stacks
	"string suffix structures":  {"q7uda9h3jl6711deg0i76fatdc"},                               // String suffix structures
	"strings":                   {"ql34pmh9fh0ofdsb8jo3brsk1s"},                               // Strings
	"suffix array":              {"9tvaiar64t2f33rlucma58ckn4"},                               // Suffix Array
	"sweep line":                {"f73kkr3fo932d0a0orkg8m4tb4"},                               // Scanline / Sweep Line
	"tarjan":                    {"msh9q06gah6hpds3m0dceu3ff8"},                               // Graphs
	"ternary search":            {"qm3576n3id76901nrcanp00alk"},                               // Ternary search
	"trees":                     {"jbe4odf0rl39rc6vtvst7kuro0"},                               // Trees
	"trivial":                   {"pjjft5joql5j95u7radbchs51g"},                               // For Beginners
	"two pointers":              {"mougogmuf10i3b5gpp7ur935l0"},                               // Two pointers
	"two-pointers":              {"mougogmuf10i3b5gpp7ur935l0"},                               // Two pointers
	"very easy":                 {"pjjft5joql5j95u7radbchs51g"},                               // For Beginners
	"while":                     {"pjjft5joql5j95u7radbchs51g"},                               // For Beginners
	"xor":                       {"mnvqr2ggu10jl7r8kpmdsusqlk"},                               // Bitwise Operations
	"z-function":                {"mfucls2rs90q9be0rgeslvt61o"},                               // Z-function
}

func TopicsFromTags(tags []SpecificationTag) (topics []string) {
	unique := map[string]bool{}
	for _, tag := range tags {
		links, ok := tagMapping[strings.ToLower(tag.Value)]
		if !ok {
			continue
		}

		for _, link := range links {
			unique[link] = true
		}
	}

	for topic := range unique {
		topics = append(topics, topic)
	}

	return
}
