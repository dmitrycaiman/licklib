package unexported

// Неэспортируемая функция.
func privateFunc() int { return 111 }
func PrivateFunc() int { return privateFunc() }

// Неэспортируемая функция-переменная.
var privateFuncVar = func() int { return 222 }

func PrivateFuncVar() int { return privateFuncVar() }

// Неэспортируемые переменные.
var (
	privateIntVar    = 333
	privateStringVar = "stringVar"
)

func PrivateIntVar() int       { return privateIntVar }
func PrivateStringVar() string { return privateStringVar }

// Неэкспортируемый тип.
type privateType struct{ a, b int }

// Неэкспортируемая переменная неэкспортируемого типа.
var privateTypeVar = &privateType{444, 555}

func PrivateTypeVarA() int { return privateTypeVar.a }
func PrivateTypeVarB() int { return privateTypeVar.b }

// Экспортируемый тип.
type PublicType struct{ A, B int }

// Неэкспортируемая переменная экспортируемого типа.
var pubilcTypeVar = &PublicType{444, 555}

func PubilcTypeVar() *PublicType { return pubilcTypeVar }
