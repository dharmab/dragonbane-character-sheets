package character

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestItemUnmarshalStringOrObject(t *testing.T) {
	t.Parallel()
	// Legacy gear slots were stored as bare strings.
	var s Item
	if err := json.Unmarshal([]byte(`"Chainmail"`), &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "Chainmail" || s.Weight != 1 {
		t.Errorf("string item = %+v; want {Chainmail 1}", s)
	}
	// New format is a full object.
	var o Item
	if err := json.Unmarshal([]byte(`{"name":"Sword","weight":2,"category":"weapon","damage":"2d8"}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.Name != "Sword" || o.Weight != 2 || o.Category != CatWeapon || o.Damage != "2d8" {
		t.Errorf("object item = %+v; want Sword/2/weapon/2d8", o)
	}
}

func TestLoadLegacyGearAutoTags(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "legacy.json")
	legacy := `{
		"armor": "Chainmail",
		"helmet": "Open Helm",
		"weapons_at_hand": ["Longsword", "", "Dagger"]
	}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Armor.Name != "Chainmail" || c.Armor.Category != CatArmor {
		t.Errorf("armor = %+v; want Chainmail/armor", c.Armor)
	}
	if c.Helmet.Name != "Open Helm" || c.Helmet.Category != CatHelmet {
		t.Errorf("helmet = %+v; want Open Helm/helmet", c.Helmet)
	}
	if len(c.WeaponsAtHand) != 3 {
		t.Fatalf("weapons len = %d; want 3", len(c.WeaponsAtHand))
	}
	if c.WeaponsAtHand[0].Category != CatWeapon || c.WeaponsAtHand[2].Category != CatWeapon {
		t.Errorf("filled weapon slots should be tagged weapon: %+v", c.WeaponsAtHand)
	}
	// An empty slot stays empty and untagged.
	if c.WeaponsAtHand[1].Name != "" || c.WeaponsAtHand[1].Category != CatNone {
		t.Errorf("empty weapon slot = %+v; want empty/untagged", c.WeaponsAtHand[1])
	}
}

func TestLoadGearRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "char.json")
	c := Default()
	c.Armor = Item{Name: "Plate", Weight: 3, Category: CatArmor, ArmorRating: 6, BaneSneaking: true, BaneEvade: true}
	c.WeaponsAtHand[0] = Item{Name: "Spear", Weight: 1, Category: CatWeapon, Grip: Grip2H, Range: 4, Damage: "2d6", Durability: 5, Features: []string{"Long"}}
	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	a := got.Armor
	if a.Name != "Plate" || a.Weight != 3 || a.Category != CatArmor || a.ArmorRating != 6 || !a.BaneSneaking || !a.BaneEvade {
		t.Errorf("armor round-trip = %+v", a)
	}
	w := got.WeaponsAtHand[0]
	if w.Grip != Grip2H || w.Range != 4 || w.Damage != "2d6" || w.Durability != 5 || len(w.Features) != 1 {
		t.Errorf("weapon round-trip = %+v", w)
	}
}

func TestLoadClampsSlotWeights(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "char.json")
	legacy := `{"armor": {"name":"Cloak","weight":0,"category":"armor"}}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Armor.Weight != 1 {
		t.Errorf("armor weight = %d; want clamped to 1", c.Armor.Weight)
	}
}
