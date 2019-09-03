package fun.build4.playground.httpbin;

import com.netflix.zuul.ZuulFilter;
import com.netflix.zuul.context.RequestContext;
import com.netflix.zuul.exception.ZuulException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

@Component
@Slf4j
public class ZuulAnythingPostFilter extends ZuulFilter {

  @Override
  public String filterType() {
    return "post";
  }

  @Override
  public int filterOrder() {
    return 999;
  }

  @Override
  public boolean shouldFilter() {
    var context = RequestContext.getCurrentContext();
    var proxyName = context.get("proxy");

    if (proxyName == null || "".equals(proxyName)) {
      return false;
    }

    return "httpbin-anything".equals(proxyName);
  }

  @Override
  public Object run() throws ZuulException {
    log.info("running anything post filter...");

    var context = RequestContext.getCurrentContext();
    var resp = context.getResponse();
    resp.addHeader("X-Application", "httpbin-zuul-spring-cloud");

    return null;
  }
}
