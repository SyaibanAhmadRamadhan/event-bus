// TODO: I think this would be suitable to be implemented separately, because this is for the producer mechanism
package erabbitmq

//
//func (r *rabbitMQ) initCloseHandler() {
//	c := make(chan os.Signal, 2)
//	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
//	go func() {
//		<-c
//		log.Printf("close handler: Ctrl+C pressed in Terminal")
//		close(r.cfg.exitCh)
//	}()
//}
//
//func (r *rabbitMQ) startConfirmModeHandler() {
//	go func() {
//		confirms := make(map[uint64]struct {
//			deliveredConfirmation *amqp.DeferredConfirmation
//			message               amqp.Publishing
//		})
//		for {
//			select {
//			case <-r.cfg.exitCh:
//				exitConfirmHandler(confirms)
//				return
//			default:
//				break
//			}
//
//			countConfirmation := len(confirms)
//			if countConfirmation <= r.cfg.maxShelterConfirmation {
//				select {
//				case r.cfg.publishOkCh <- struct{}{}:
//					log.Println("permission ok to publish")
//				case <-time.After(5 * time.Second):
//					log.Println("confirm handler: timeout indicating OK to publish (this should never happen!)")
//				}
//			} else {
//				// mechanisme retry for sending message, asumsi rabbitmq error when ack message
//				log.Printf("blocking publish until confirmation <= %d", 8)
//			}
//
//			select {
//			case confirmation := <-r.cfg.confirmsCh:
//				deliveryTag := confirmation.deliveredConfirmation.DeliveryTag
//				confirms[deliveryTag] = confirmation
//			case <-r.cfg.exitCh:
//				exitConfirmHandler(confirms)
//				return
//			}
//
//			checkConfirmations(confirms)
//		}
//	}()
//}
//
//func exitConfirmHandler(confirms map[uint64]struct {
//	deliveredConfirmation *amqp.DeferredConfirmation
//	message               amqp.Publishing
//}) {
//	log.Println("confirm handler: exit requested")
//	waitingConfirmation(confirms)
//	log.Println("confirm handler: exiting")
//}
//
//func checkConfirmations(confirms map[uint64]struct {
//	deliveredConfirmation *amqp.DeferredConfirmation
//	message               amqp.Publishing
//}) {
//	log.Printf("confirm handler: checking %d outstanding confirmations", len(confirms))
//	for k, v := range confirms {
//		v.deliveredConfirmation.Done()
//		if v.deliveredConfirmation.Acked() {
//			log.Printf("confirm handler: confirmed delivery with tag: %d", k)
//			delete(confirms, k)
//		}
//	}
//}
//
//func waitingConfirmation(confirms map[uint64]struct {
//	deliveredConfirmation *amqp.DeferredConfirmation
//	message               amqp.Publishing
//}) {
//	log.Println("confirm handler: waiting confirmation")
//
//	checkConfirmations(confirms)
//
//	// todo retry again for ack delivered message
//	outstandingConfirmation := len(confirms)
//	log.Println("confirm handler: outstanding confirmation: %d", outstandingConfirmation)
//}
