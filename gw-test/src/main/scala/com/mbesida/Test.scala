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
    Future.sequence{
      (0 to 10).map { i =>
        val start = System.currentTimeMillis()
        client.sendAsync(
          HttpRequest.newBuilder().GET().uri(URI.create("http://localhost:8000/get-fortune")).build(),
          BodyHandlers.ofString()
        ).asScala.map( httpResponse =>
          (i, httpResponse.statusCode(), httpResponse.body(), System.currentTimeMillis() - start)
        )
    }
  }

  val start = System.currentTimeMillis()
  Await.result(result, Duration.Inf).foreach{ (i, status, body, duration) =>
    println(s"$i - $duration ms - status $status - $body")
  }
  println(s"Took ${System.currentTimeMillis() - start} ms")
