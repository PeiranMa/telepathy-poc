#!/usr/bin/env python
import time
import threading

from confluent_kafka import Producer, Consumer
import sys
import config
import logging
from metrics import profile

from confluent_kafka.admin import AdminClient, NewTopic

FORMAT = '%(asctime)-15s %(message)s'
logger = logging.getLogger("producer")
logger.setLevel(logging.INFO)
logging.basicConfig(filename="producer.log", filemode="w", format=FORMAT)


class ProducerClient:

    def __init__(self):
        conf = {'bootstrap.servers': config.BOOTSTRAP_SERVER}
        self._producer = Producer(**conf)
        consumer_conf = {'bootstrap.servers': config.BOOTSTRAP_SERVER, 'session.timeout.ms': 6000,
            'auto.offset.reset': 'earliest', 'group.id': 1001,}
        self._consumer =  Consumer(consumer_conf)

    @profile(logger)
    def submit_job(self, job):
        logger.debug("submit job start " + str(job.jobid))

        

        a = AdminClient({'bootstrap.servers': config.BOOTSTRAP_SERVER})

        new_topics = [NewTopic(topic, num_partitions=config.PARTITION_NUM, replication_factor=1) for topic in [config.JOB_SUBMIT_TOPIC + str(job.jobid)]]

        fs = a.create_topics(new_topics)
        for topic, f in fs.items():
            f.result()

        def delivery_callback(err, msg):
            if err:
                print("{0} deliver error: {1}".format(msg, err))

        for task in job.tasks:
            try:
            # Produce line (without newline)
                self._producer.produce(config.JOB_SUBMIT_TOPIC + str(job.jobid), str(task.taskid), partition=task.taskid % 32, timestamp=int(task.timestamp), callback=delivery_callback)

            except BufferError:
                logger.debug('%% Local producer queue is full (%d messages awaiting delivery): try again\n' %
                                len(self._producer))
            self._producer.poll(0)

        logger.info('%% Waiting for %d deliveries\n' % len(self._producer))
        self._producer.flush()
    
    @profile(logger=logger)
    def monitor_job(self, job):
        max_cost = -1
        cnt = 0
        now = time.time()
        self._consumer.subscribe([config.JOB_FINISH_TOPIC + str(config.JOB_ID)])
        while True:
            msg = self._consumer.poll(timeout=1.0)
            if msg is None:
                continue
            if msg.error():
                raise Exception(msg.error())
            else:
                # taskid = int(msg.value())
                _, finish_time = msg.timestamp()
                # max_cost = max(max_cost, finish_time - job.tasks[taskid].timestamp)
                cnt += 1
                if cnt % 10000 == 0:
                    logger.info("monitor {0} tasks cost {1} sec".format(cnt, time.time() - now))
                    # logger.info("max_cost {0} sec".format(max_cost))
                if cnt >= len(job.tasks) * 0.9:
                    # logger.info("finish job cost:" + str(max_cost))
                    return

        

if __name__ == "__main__":
    from job import Job, Task

    def main_producer():
        job = Job(config.JOB_ID)
        for i in range(config.TASK_NUM):
            job.tasks.append(Task(i))
        
        p = ProducerClient()
        t1 = threading.Thread(target=p.submit_job, args=(job,))
        t1.start()
        p.monitor_job(job)

    main_producer()

