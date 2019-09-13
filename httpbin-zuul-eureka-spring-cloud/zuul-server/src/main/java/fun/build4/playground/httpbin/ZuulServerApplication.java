package fun.build4.playground.httpbin;

import org.springframework.boot.Banner.Mode;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.eureka.EnableEurekaClient;
import org.springframework.cloud.netflix.zuul.EnableZuulProxy;

@SpringBootApplication
@EnableEurekaClient
@EnableZuulProxy
public class ZuulServerApplication {

  public static void main(String[] args) {
    var app = new SpringApplication(ZuulServerApplication.class);
    app.setBannerMode(Mode.OFF);
    app.run(args);
  }
}
