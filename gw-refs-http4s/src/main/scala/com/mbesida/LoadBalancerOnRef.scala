package com.mbesida

import org.http4s.Request
import org.http4s.Response
import org.http4s.client.Client
import org.http4s.Uri
import cats.effect.*
import scala.concurrent.duration.*

trait LoadBalancerOnRef:
  def loadBalance(req: Request[IO]): IO[Response[IO]]

object LoadBalancerOnRef:

  private val SleepDuration = 1500.millis

  private def lockService(busyMap: Map[Uri, Boolean]): (Map[Uri, Boolean], Option[Uri]) =
    busyMap
      .find((_, isBusy) => !isBusy)
      .map((u, _) => (busyMap.updated(u, true), Some(u)))
      .getOrElse((busyMap, None))

  def apply(client: Client[IO], serviceUrls: Seq[Uri]): IO[LoadBalancerOnRef] =

    val mutexIO = IO.ref(
      serviceUrls.zip(List.fill(serviceUrls.length)(false)).toMap
    )

    mutexIO.map { busyMap =>
      new:
        def loadBalance(req: Request[IO]): IO[Response[IO]] = {
          busyMap.modify(m => lockService(m)).flatMap {
            case None => IO.sleep(SleepDuration) >> loadBalance(req)
            case Some(uri) =>
              client.run(req.withUri(uri)).use { resp =>
                busyMap.update(_.updated(uri, false)).uncancelable *> {
                  if resp.status.isSuccess then IO.pure(resp)
                  else IO.sleep(SleepDuration) >> loadBalance(req)
                }
              }
          }
        }
    }
