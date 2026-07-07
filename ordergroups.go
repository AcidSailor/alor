package alor

import (
	"context"
	"net/http"
	"net/url"

	"github.com/acidsailor/restkit"
)

// orderGroupsService groups the OrderGroups API operations. Obtain it via Client.OrderGroups.
type orderGroupsService struct{ c *Client }

// List returns all order groups (linked-order baskets).
func (s *orderGroupsService) List(
	ctx context.Context,
) ([]ResponseOrderGroupInfo, error) {
	return do[[]ResponseOrderGroupInfo](ctx, s.c, http.MethodGet,
		"/commandapi/api/orderGroups", restkit.NewValues(), nil)
}

// OrderGroupsGetRequest selects the single order group to return.
type OrderGroupsGetRequest struct {
	OrderGroupID string `json:"orderGroupId"`
}

// Get returns a single order group by id.
func (s *orderGroupsService) Get(
	ctx context.Context,
	params OrderGroupsGetRequest,
) (*ResponseOrderGroupInfo, error) {
	return do[*ResponseOrderGroupInfo](ctx, s.c, http.MethodGet,
		"/commandapi/api/orderGroups/"+url.PathEscape(params.OrderGroupID),
		restkit.NewValues(), nil)
}

// OrderGroupsCreateRequest carries the order group to create.
type OrderGroupsCreateRequest struct {
	Group OrderGroupCreate `json:"group"`
}

// Create links the given orders into a new order group and returns its
// id.
func (s *orderGroupsService) Create(
	ctx context.Context,
	params OrderGroupsCreateRequest,
) (*ResponseOrderGroupCreationSuccess, error) {
	return do[*ResponseOrderGroupCreationSuccess](ctx, s.c, http.MethodPost,
		"/commandapi/api/orderGroups", restkit.NewValues(), params.Group)
}

// OrderGroupsUpdateRequest selects the order group to modify and carries the
// modification.
type OrderGroupsUpdateRequest struct {
	OrderGroupID string           `json:"orderGroupId"` // path
	Group        OrderGroupModify `json:"group"`
}

// Update modifies the order group with the given id. Returns only an
// error (bare success reply).
func (s *orderGroupsService) Update(
	ctx context.Context,
	params OrderGroupsUpdateRequest,
) error {
	path := "/commandapi/api/orderGroups/" + url.PathEscape(params.OrderGroupID)
	return exec(
		ctx,
		s.c,
		http.MethodPut,
		path,
		restkit.NewValues(),
		params.Group,
	)
}

// OrderGroupsDeleteRequest selects the order group to delete.
type OrderGroupsDeleteRequest struct {
	OrderGroupID string `json:"orderGroupId"` // path
}

// Delete deletes the order group with the given id. Returns only an
// error (bare success reply).
func (s *orderGroupsService) Delete(
	ctx context.Context,
	params OrderGroupsDeleteRequest,
) error {
	path := "/commandapi/api/orderGroups/" + url.PathEscape(params.OrderGroupID)
	return exec(ctx, s.c, http.MethodDelete, path, restkit.NewValues(), nil)
}
