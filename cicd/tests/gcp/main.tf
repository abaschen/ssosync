provider "googleworkspace" {
  credentials             = var.google_credentialsJson
  impersonated_user_email = var.google_impersonatedEmail
  customer_id             = var.google_customerId
  oauth_scopes = [
    "https://www.googleapis.com/auth/admin.directory.user",
    "https://www.googleapis.com/auth/admin.directory.group",
    # include scopes as needed
  ]
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}


resource "googleworkspace_user" "alice" {
  primary_email = "alice@${var.google_domain}"
  name {
    given_name  = "Alice"
    family_name = "Smith"
  }
  password = random_password.password.result
}

resource "googleworkspace_user" "bob" {
  primary_email = "bob@${var.google_domain}"
  name {
    given_name  = "Bob"
    family_name = "Jones"
  }
  password = random_password.password.result
}


resource "googleworkspace_group" "aws_admins" {
  email       = "aws-admins@${var.google_domain}"
  name        = "AWS Admins"
  description = "Group for AWS Administrators"
}

resource "googleworkspace_group" "aws_devs" {
  email       = "aws-devs@${var.google_domain}"
  name        = "AWS Developers"
  description = "Group for AWS Developers"
}

resource "googleworkspace_group_member" "alice_admin" {
  group_id = googleworkspace_group.aws_admins.id
  email    = googleworkspace_user.alice.primary_email
  role     = "MEMBER"
}

resource "googleworkspace_group_member" "bob_dev" {
  group_id = googleworkspace_group.aws_devs.id
  email    = googleworkspace_user.bob.primary_email
  role     = "MEMBER"
}

