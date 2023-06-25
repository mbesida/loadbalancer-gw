package com.mbesida

import org.http4s.Request
import org.http4s.Response
import org.http4s.client.Client
import java.util.concurrent.Semaphore
import org.http4s.Uri
import cats.effect.*
import scala.concurrent.duration.*

trait LoadBalancer:
  def loadBalance(req: Request[IO]): IO[Response[IO]]

object LoadBalancer {

  private val SleepDuration = 1500.millis

  def apply(client: Client[IO], serviceUrls: Seq[Uri]): LoadBalancer = {

    val mutexes =
      serviceUrls.zip(List.fill(serviceUrls.length)(new Semaphore(1)))

    new:
      def loadBalance(req: Request[IO]): IO[Response[IO]] = {
        mutexes.find((_, m) => m.tryAcquire()) match
          case None => IO.sleep(SleepDuration) >> loadBalance(req)
          case Some((uri, mutex)) =>
            client.run(req.withUri(uri)).use { resp =>
              IO(mutex.release()).uncancelable *> {
                if resp.status.isSuccess then IO.pure(resp)
                else IO.sleep(SleepDuration) >> loadBalance(req)
              }
            }
      }
  }
}
