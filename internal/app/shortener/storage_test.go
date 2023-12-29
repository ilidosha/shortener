package shortener

import "testing"

func TestStorage_Append(t *testing.T) {
	type fields struct {
		Records map[string]string
	}
	type args struct {
		url string
	}
	m := make(map[string]string)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "works",
			fields:  fields{m},
			args:    args{url: "http://example.org/"},
			want:    generateShortKey(),
			wantErr: false,
		},
		{
			name:    "is not a valid url",
			fields:  fields{m},
			args:    args{url: "http//example.org/"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "return unexpected error, cause provided no fields to insert in them",
			fields:  fields{},
			args:    args{url: "http://example.org/"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				Records: tt.fields.Records,
			}
			got, err := s.Append(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Append() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// just checking length of the strings cause i don't know how to properly check it (same type and length or something)
			if len(got) != len(tt.want) {
				t.Errorf("Append() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateShortKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Works",
			want: "kBG4Fo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// just checking length of the strings cause i don't know how to properly check it
			if got := generateShortKey(); len(got) != len(tt.want) {
				t.Errorf("generateShortKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
