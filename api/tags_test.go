package api

import (
	"reflect"
	"testing"
)

func Test_tags_inOrg(t *testing.T) {
	type fields struct {
		org  string
		tags Tags
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test matching org",
			fields: fields{
				org: "Metaverse",
				tags: Tags{
					{Key: "spinup:org", Value: "Metaverse"},
				},
			},
			want: true,
		},
		{
			name: "test conflicting org",
			fields: fields{
				org: "Metaverse",
				tags: Tags{
					{Key: "spinup:org", Value: "Fedland"},
				},
			},
			want: false,
		},
		{
			name: "test missing org",
			fields: fields{
				org: "Metaverse",
				tags: Tags{
					{Key: "CreatedBy", Value: "Hiro"},
				},
			},
			want: false,
		},
		{
			name: "test empty tags",
			fields: fields{
				org:  "Metaverse",
				tags: nil,
			},
			want: false,
		},
		{
			name: "test empty tags blank org",
			fields: fields{
				org:  "",
				tags: nil,
			},
			want: false,
		},
		{
			name: "test blank org match",
			fields: fields{
				org: "",
				tags: Tags{
					{Key: "spinup:org", Value: ""},
				},
			},
			want: true,
		},
		{
			name: "test blank org mismatch",
			fields: fields{
				org: "",
				tags: Tags{
					{Key: "spinup:org", Value: "Fedland"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := tt.fields.tags
			got := tags.inOrg(tt.fields.org)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tags.inOrg()\ngot:  %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func Test_tags_normalize(t *testing.T) {
	type fields struct {
		org  string
		tags Tags
	}

	tests := []struct {
		name   string
		fields fields
		want   Tags
	}{
		{
			name: "test conflicting org",
			fields: fields{
				org: "testOrg",
				tags: Tags{
					{Key: "spinup:org", Value: "XOR"},
				},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
			},
		},
		{
			name: "test conflicting type",
			fields: fields{
				org: "testOrg",
				tags: Tags{
					{Key: "spinup:type", Value: "container"},
				},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
			},
		},
		{
			name: "test conflicting flavor",
			fields: fields{
				org: "testOrg",
				tags: Tags{
					{Key: "spinup:flavor", Value: "task"},
				},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
			},
		},
		{
			name: "test multiple conflicting tags",
			fields: fields{
				org: "testOrg",
				tags: Tags{
					{Key: "spinup:org", Value: "XOR"},
					{Key: "spinup:type", Value: "container"},
					{Key: "spinup:flavor", Value: "task"},
				},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
			},
		},
		{
			name: "test preserving user tags",
			fields: fields{
				org: "testOrg",
				tags: Tags{
					{Key: "CreatedBy", Value: "me"},
					{Key: "Env", Value: "test"},
				},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
				{Key: "CreatedBy", Value: "me"},
				{Key: "Env", Value: "test"},
			},
		},
		{
			name: "test setting org and preserving user tags",
			fields: fields{
				org: "testOrg",
				tags: Tags{
					{Key: "spinup:org", Value: "Rogue"},
					{Key: "CreatedBy", Value: "me"},
					{Key: "Env", Value: "test"},
				},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
				{Key: "CreatedBy", Value: "me"},
				{Key: "Env", Value: "test"},
			},
		},
		{
			name: "test empty tags",
			fields: fields{
				org:  "testOrg",
				tags: Tags{},
			},
			want: Tags{
				{Key: "spinup:org", Value: "testOrg"},
				{Key: "spinup:type", Value: "storage"},
				{Key: "spinup:flavor", Value: "datamover"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := tt.fields.tags
			got := tags.normalize(tt.fields.org)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tags.normalize()\ngot:  %v\nwant: %v", got, tt.want)
			}
		})
	}
}
