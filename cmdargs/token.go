package cmdargs

type Role int

func (r Role) Has(role Role) bool {
	return r&role != 0
}

const (
	RoleFlag       Role = 1 << iota
	RoleBoolFlag        = 1 << iota // modifies RoleFlag
	RoleKnown           = 1 << iota // modifies RoleFlag or RoleFlagValue
	RoleInline          = 1 << iota // modifies RoleFlag
	RoleFlagValue       = 1 << iota
	RoleUnnamed         = 1 << iota
	RoleTerminator      = 1 << iota
)

type Token struct {
	Arg       string
	FlagName  string
	FlagValue string
	// Role is sum of Role constants. Possible values:
	// RoleFlag | RoleKnown | RoleInline                 // contains FlagValue
	// RoleFlag | RoleKnown | RoleBoolFlag               // no FlagValue, implicit `true`
	// RoleFlag | RoleKnown | RoleInline | RoleBoolFlag  // explicit bool string FlagValue
	// RoleFlag | RoleKnown                              // will be followed by value
	// RoleFlag | RoleInline                             // unknown, contains FlagValue
	// RoleFlag | RoleBoolFlag                           // unknown, located at the end, no FlagValue
	// RoleFlag                 //  unknown, it's ambiguous whether it's bool flag or flag name
	//                          //  On receiving it, yield function should decide how to treat it
	// RoleUnnamed
	// RoleFlagValue            // goes next after non-inline flag
	// RoleTerminator
	Role Role
}
