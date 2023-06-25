ThisBuild / scalaVersion := "3.3.0"

lazy val http4sVersion = "0.23.20"
lazy val http4sServer = "org.http4s" %% "http4s-ember-server" % http4sVersion
lazy val http4sDsl = "org.http4s" %% "http4s-dsl" % http4sVersion
lazy val http4sClient = "org.http4s" %% "http4s-ember-client" % http4sVersion

lazy val balancer = project
  .in(file("."))
  .aggregate(`gw-refs-http4s`, `gw-semaphore-http4s`)

lazy val `gw-refs-http4s` = project
  .in(file("gw-refs-http4s"))
  .settings(
    name := "gw-refs-http4s"
  )

lazy val `gw-semaphore-http4s` = project
  .in(file("gw-semaphore-http4s"))
  .settings(
    name := "gw-semaphore-http4s",
    libraryDependencies := Seq(http4sClient, http4sDsl, http4sServer)
  )
