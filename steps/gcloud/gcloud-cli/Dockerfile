FROM us-docker.pkg.dev/stackpulse/public/gcloud/base
ARG STEP_BASEPATH
COPY ${STEP_BASEPATH}/gcloud-entrypoint.py /tmp/
ENTRYPOINT ["python", "/tmp/gcloud-entrypoint.py"]