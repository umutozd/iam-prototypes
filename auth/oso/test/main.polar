allow(actor, action, resource) if
  has_permission(actor, action, resource);

actor User {}

resource Ticket {
	permissions = ["read", "write", "delete"];
	roles = [
        "otsimo.com/ticket/viewer",
        "otsimo.com/ticket/editor",
        "otsimo.com/ticket/maintainer"
    ];

	"read" if "otsimo.com/ticket/viewer";
	"write" if "otsimo.com/ticket/editor";
	"delete" if "otsimo.com/ticket/maintainer";

	"otsimo.com/ticket/editor" if "otsimo.com/ticket/maintainer";
	"otsimo.com/ticket/viewer" if "otsimo.com/ticket/editor";
}

has_role(user: User, roleName: String, _: Ticket) if
  role in user.Roles and
  role = roleName;
