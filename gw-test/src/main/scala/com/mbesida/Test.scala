package com.mbesida

import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.URI
import java.net.http.HttpResponse.BodyHandlers
import scala.concurrent.ExecutionContext.Implicits.global
import scala.jdk.FutureConverters.*
import scala.concurrent.duration.*
import scala.concurrent.*

object Test extends App:
  val client = HttpClient.newHttpClient()

  val result =
    Future.sequence {
      (0 to 10).map { i =>
        val start = System.currentTimeMillis()
        client
          .sendAsync(
            HttpRequest
              .newBuilder()
              .GET()
              .uri(URI.create("http://localhost:8000/get-fortune"))
              .build(),
            BodyHandlers.ofString()
          )
          .asScala
          .map{ httpResponse =>
            val (status, duration, body) = (httpResponse.statusCode(), System.currentTimeMillis() - start, httpResponse.body())
            println(s"$i - $duration ms - status $status - $body")
          }
      }
    }

  val start = System.currentTimeMillis()
  Await.result(result, Duration.Inf)
  println(s"Took ${System.currentTimeMillis() - start} ms")
