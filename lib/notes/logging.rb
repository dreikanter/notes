module Notes::Logging
  def logger
    @logger ||= Logger.new(stream).tap { _1.formatter = formatter }
  end

  private

  def stream
    ENV["SUPPRESS_LOGS"] ? "/dev/null" : $stdout
  end

  def formatter
    proc { |_severity, _datetime, _progname, msg| "#{msg}\n" }
  end
end
