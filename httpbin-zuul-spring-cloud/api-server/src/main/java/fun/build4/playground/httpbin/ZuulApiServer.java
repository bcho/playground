package fun.build4.playground.httpbin;

import java.util.HashMap;
import java.util.Map;
import javax.servlet.http.HttpServletRequest;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.Banner.Mode;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.lang.NonNull;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@SpringBootApplication
@RestController
@Slf4j
public class ZuulApiServer {

  public static void main(String[] args) {
    var app = new SpringApplication(ZuulApiServer.class);
    app.setBannerMode(Mode.OFF);
    app.run(args);
  }

  @RequestMapping({"/anything", "/anything/**"})
  public AnythingResponse anything(HttpServletRequest req) {
    return AnythingResponse.fromReq(req);
  }

  @Data
  @NoArgsConstructor
  @AllArgsConstructor
  @Builder
  private static final class AnythingResponse {

    @NonNull private String method;
    @NonNull private String url;
    @NonNull private String origin;
    @NonNull private Map<String, String> headers;

    static AnythingResponse fromReq(@NonNull HttpServletRequest req) {
      return AnythingResponse.builder()
          .method(req.getMethod())
          .url(req.getRequestURL().toString())
          .origin(getOriginFromRequest(req))
          .headers(getHeadersFromRequest(req))
          .build();
    }

    static String getOriginFromRequest(@NonNull HttpServletRequest req) {
      var xForwardedFor = req.getHeader("X-Forwarded-For");
      if (xForwardedFor != null && !"".equals(xForwardedFor)) {
        return xForwardedFor;
      }

      return req.getRemoteAddr();
    }

    static Map<String, String> getHeadersFromRequest(@NonNull HttpServletRequest req) {
      final Map<String, String> rv = new HashMap<>();

      var headerNames = req.getHeaderNames();
      while (headerNames.hasMoreElements()) {
        var key = headerNames.nextElement();
        rv.put(key, req.getHeader(key));
      }

      return rv;
    }
  }
}
