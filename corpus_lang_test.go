package main

import (
	"testing"

	"golang.org/x/text/language"
)

func TestGetLangCorpus(t *testing.T) {
	save := make([]*Corpus, 100)
	type args struct {
		language language.Tag
		wordSize int
	}
	tests := []struct {
		name    string
		args    args
		save    int
		want    int
		wantErr bool
	}{
		{"da_5_1", args{language.Danish, 5}, 1, 0, false},
		{"da_5_2", args{language.Danish, 5}, 0, 1, false},
		{"da_6_1", args{language.Danish, 6}, 2, 0, false},
		{"da_5_3", args{language.Danish, 5}, 0, 1, false},
		{"da_6_2", args{language.Danish, 6}, 0, 2, false},
		{"da_1_1", args{language.Danish, 1}, 11, 0, false},
		{"da_2_1", args{language.Danish, 2}, 12, 0, false},
		{"da_3_1", args{language.Danish, 3}, 13, 0, false},
		{"da_4_1", args{language.Danish, 4}, 14, 0, false},
		{"da_7_1", args{language.Danish, 7}, 17, 0, false},
		{"da_8_1", args{language.Danish, 8}, 18, 0, false},
		{"da_9_1", args{language.Danish, 9}, 19, 0, false},
		{"da_5_4", args{language.Danish, 5}, 0, 1, false},
		{"da_10_1", args{language.Danish, 10}, 20, 0, false},
		{"da_11_1", args{language.Danish, 11}, 21, 0, false},
		{"da_12_1", args{language.Danish, 12}, 22, 0, false},
		{"da_13_1", args{language.Danish, 13}, 23, 0, false},
		{"da_14_1", args{language.Danish, 14}, 24, 0, false},
		{"da_15_1", args{language.Danish, 15}, 25, 0, false},
		{"da_16_1", args{language.Danish, 16}, 26, 0, false},
		{"da_6x_1", args{language.Danish, 6}, 56, 0, false},
		{"da_6x_2", args{language.Danish, 6}, 0, 56, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLangCorpus(tt.args.language, tt.args.wordSize)
			if tt.save > 0 {
				save[tt.save] = got
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLangCorpus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want > 0 && save[tt.want] != got {
				t.Errorf("GetLangCorpus() = %p, want [%d] %p", got, tt.want, save[tt.want])
			}
			for i := range save {
				if i > 0 && i != tt.want && i != tt.save && save[i] == got {
					if tt.want > 0 {
						t.Errorf("GetLangCorpus() = got [%d] %p but this equals [%d] %p ", tt.want, got, i, save[i])
					} else {
						t.Errorf("GetLangCorpus() = got %p not expected to be found but this equals [%d] %p ", got, i, save[i])
					}
				}
			}
		})
	}
}
