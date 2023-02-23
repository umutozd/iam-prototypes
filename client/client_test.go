package client

import (
	"context"
	"testing"

	"github.com/umutozd/iam-prototypes/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestClientCreds(t *testing.T) {

	cases := []struct {
		name                string
		token               string
		expectedGetError    error
		expectedUpdateError error
	}{
		{
			name:                "token-with-viewer-group",
			token:               "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NzcxNTI2OTAsImV4cCI6MTcwODY4ODY5MCwiYXVkIjoiIiwic3ViIjoiZDVkMzk0MmItZTUyZi00ZjdkLWE3OWUtZDc0ZGI2MzQwMjNjIiwiZ3JvdXBzIjpbIm90c2ltby5jb20vZm9vL3ZpZXdlciIsIm90c2ltby5jb20vZGV2ZWxvcGVyIl19.W-beD2XbtzRCvFt7pU5HjGFkiCYhvMRu7o8zwveLAEY",
			expectedGetError:    nil,
			expectedUpdateError: status.Error(codes.PermissionDenied, "permission denied"),
		},
		{
			name:                "token-with-viewer-and-editor-groups",
			token:               "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NzcxNTI2OTAsImV4cCI6MTcwODY4ODY5MCwiYXVkIjoiIiwic3ViIjoiZDVkMzk0MmItZTUyZi00ZjdkLWE3OWUtZDc0ZGI2MzQwMjNjIiwiZ3JvdXBzIjpbIm90c2ltby5jb20vZm9vL3ZpZXdlciIsIm90c2ltby5jb20vZm9vL2VkaXRvciJdfQ.uFUJKsyvtMlG6u8hKVaFifRDL2UlhF_Clv3ipnTzRRo",
			expectedGetError:    nil,
			expectedUpdateError: nil,
		},
		{
			name:                "token-with-maintainer-group",
			token:               "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NzcxNTI2OTAsImV4cCI6MTcwODY4ODY5MCwiYXVkIjoiIiwic3ViIjoiZDVkMzk0MmItZTUyZi00ZjdkLWE3OWUtZDc0ZGI2MzQwMjNjIiwiZ3JvdXBzIjpbIm90c2ltby5jb20vZGV2ZWxvcGVyIiwib3RzaW1vLmNvbS9mb28vbWFpbnRhaW5lciJdfQ.ZoewPES9mpoOzCSLni0Qr3iWme6WA2kwviqpNzrQj9E",
			expectedGetError:    nil,
			expectedUpdateError: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// create client connection
			conn, err := grpc.Dial(":8080",
				grpc.WithBlock(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithPerRPCCredentials(&perRPCCredsWithToken{token: c.token}),
			)
			if err != nil {
				t.Fatalf("error dialing server: %v", err)
			}
			cli := pb.NewSimpleServiceClient(conn)

			// test GetFoo
			_, err = cli.GetFoo(context.Background(), &pb.GetFooReq{Name: "foo-1"})
			_ = compareTestErrors(t, c.expectedGetError, err)

			// test UpdateFoo
			_, err = cli.UpdateFoo(context.Background(), &pb.UpdateFooReq{FooId: "foo-1", Count: 42})
			_ = compareTestErrors(t, c.expectedUpdateError, err)
		})
	}
}

func compareTestErrors(t *testing.T, expected, got error) (hasError bool) {
	if expected != nil {
		if got == nil {
			t.Fatalf("got nil error, but expected: %v", expected)
		} else if expected.Error() != got.Error() {
			t.Fatalf("wrong errors: expected=%v, got=%v", expected, got)
		} else {
			return true
		}
	} else {
		if got != nil {
			t.Fatalf("expected nil error, but got: %v", got)
		}
	}
	return false
}

type perRPCCredsWithToken struct {
	token string
}

func (c *perRPCCredsWithToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer" + " " + c.token,
	}, nil
}

func (c *perRPCCredsWithToken) RequireTransportSecurity() bool {
	return false
}
