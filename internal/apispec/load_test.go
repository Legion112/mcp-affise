package apispec_test
import (
  "testing"
  "github.com/Legion112/mcp-affise/internal/apispec"
)
func TestLoadEmbedded(t *testing.T) {
  c, err := apispec.Load()
  if err != nil { t.Fatal(err) }
  if len(c.Endpoints) < 80 { t.Fatalf("got %d", len(c.Endpoints)) }
  _, _, err = c.Resolve("/3.0/offers")
  if err != nil { t.Fatal(err) }
}
