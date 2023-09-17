package com.mbesida

import com.comcast.ip4s.*
import cats.implicits.*
import cats.effect.*
import org.http4s.*
import org.http4s.dsl.io.*
import org.http4s.implicits.*
import org.http4s.ember.server.EmberServerBuilder
import org.http4s.ember.client.EmberClientBuilder

object Server extends IOApp.Simple {

  val workers = List(
    uri"http://localhost:9551",
    uri"http://localhost:9552",
    uri"http://localhost:9553"
  ).map(u => u / "get-fortune")

  def service(balancer: LoadBalancer) = HttpRoutes.of[IO]:
    case req @ GET -> Root / "get-fortune" => balancer.loadBalance(req)

  override def run: IO[Unit] =
    EmberClientBuilder
      .default[IO]
      .build
      .flatMap { client =>
        EmberServerBuilder
          .default[IO]
          .withHost(ipv4"0.0.0.0")
          .withPort(port"8000")
          .withHttpApp(service(LoadBalancer(client, workers)).orNotFound)
          .build
      }
      .use { _ =>
        IO.println("Load balancer has started") *> IO.never
      }
}
