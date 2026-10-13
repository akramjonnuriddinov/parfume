package assistant

import (
	"slices"
	"testing"
)

func TestExtractTrailingOptions(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		wantText  string
		wantReply []string
	}{
		{
			name:      "options after question",
			in:        "Savatga qo‘shildi. Davom etamizmi?\n\n- Buyurtma berish\n- Yana mahsulot ko‘rish",
			wantText:  "Savatga qo‘shildi. Davom etamizmi?",
			wantReply: []string{"Buyurtma berish", "Yana mahsulot ko‘rish"},
		},
		{
			name:     "content list is kept",
			in:       "Sizga mos:\n- Tonometr — bosim uchun\n- Termometr — harorat uchun\nTanlang.",
			wantText: "Sizga mos:\n- Tonometr — bosim uchun\n- Termometr — harorat uchun\nTanlang.",
		},
		{
			name:     "list without blank line is kept",
			in:       "Nima kerak?\n- Bosim\n- Qand",
			wantText: "Nima kerak?\n- Bosim\n- Qand",
		},
		{
			name:     "single bullet is kept",
			in:       "Davom etamizmi?\n\n- Ha",
			wantText: "Davom etamizmi?\n\n- Ha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &turn{}
			got := extractTrailingOptions(tr, tt.in)
			if got != tt.wantText {
				t.Errorf("text = %q, want %q", got, tt.wantText)
			}
			if !slices.Equal(tr.reply.QuickReplies, tt.wantReply) {
				t.Errorf("quick replies = %q, want %q", tr.reply.QuickReplies, tt.wantReply)
			}
		})
	}
}

func TestExtractLeakedReplies(t *testing.T) {
	tr := &turn{}
	got := extractLeakedReplies(tr, "Qo‘shildi.\n\nsuggest_replies: [\"Ha\", \"Yo‘q\"]")
	if got != "Qo‘shildi." {
		t.Errorf("text = %q", got)
	}
	if !slices.Equal(tr.reply.QuickReplies, []string{"Ha", "Yo‘q"}) {
		t.Errorf("quick replies = %q", tr.reply.QuickReplies)
	}
}

func TestExtractLeakedRepliesBullets(t *testing.T) {
	tr := &turn{}
	got := extractLeakedReplies(tr, "Davom etamizmi?\n\nsuggest_replies:\n- Ha, menejer bog'lansin\n- Savolim bor")
	if got != "Davom etamizmi?" {
		t.Errorf("text = %q", got)
	}
	if !slices.Equal(tr.reply.QuickReplies, []string{"Ha, menejer bog'lansin", "Savolim bor"}) {
		t.Errorf("quick replies = %q", tr.reply.QuickReplies)
	}
}

func TestMoney(t *testing.T) {
	for in, want := range map[int64]string{0: "0 so‘m", 950: "950 so‘m", 1250000: "1 250 000 so‘m", 100000: "100 000 so‘m"} {
		if got := money(in); got != want {
			t.Errorf("money(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestOrderNumber(t *testing.T) {
	if got := OrderNumber(42); got != "#00042" {
		t.Errorf("OrderNumber(42) = %q", got)
	}
}
