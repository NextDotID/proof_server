package discord

import (
	"github.com/nextdotid/proof-server/validator"
	"reflect"
	"testing"
)

func TestDiscord_GeneratePostPayload(t *testing.T) {
	type fields struct {
		Base *validator.Base
	}
	tests := []struct {
		name     string
		fields   fields
		wantPost map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &Discord{
				Base: tt.fields.Base,
			}
			if gotPost := dc.GeneratePostPayload(); !reflect.DeepEqual(gotPost, tt.wantPost) {
				t.Errorf("GeneratePostPayload() = %v, want %v", gotPost, tt.wantPost)
			}
		})
	}
}

func TestDiscord_GenerateSignPayload(t *testing.T) {
	type fields struct {
		Base *validator.Base
	}
	tests := []struct {
		name        string
		fields      fields
		wantPayload string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &Discord{
				Base: tt.fields.Base,
			}
			if gotPayload := dc.GenerateSignPayload(); gotPayload != tt.wantPayload {
				t.Errorf("GenerateSignPayload() = %v, want %v", gotPayload, tt.wantPayload)
			}
		})
	}
}

func TestDiscord_Validate(t *testing.T) {
	type fields struct {
		Base *validator.Base
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &Discord{
				Base: tt.fields.Base,
			}
			if err := dc.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscord_validateText(t *testing.T) {
	type fields struct {
		Base *validator.Base
	}
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &Discord{
				Base: tt.fields.Base,
			}
			if err := dc.validateText(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("validateText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
		})
	}
}
