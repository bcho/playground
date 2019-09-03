package fun.build4.playground.httpbin;

import org.springframework.boot.Banner.Mode;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.zuul.EnableZuulProxy;

@SpringBootApplication(scanBasePackages = {"fun.build4.playground.httpbin"})
@EnableZuulProxy
public class ZuulEdgeServer {

  public static void main(String[] args) {
    var app = new SpringApplication(ZuulEdgeServer.class);
    app.setBannerMode(Mode.OFF);
    app.run(args);
  }
}
